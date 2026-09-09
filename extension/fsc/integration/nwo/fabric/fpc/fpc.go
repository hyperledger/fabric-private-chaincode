/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/fabricconfig"
	nnetwork "github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/network"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/packager"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/topology"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/services/logging"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/lifecycle"
	"github.com/hyperledger/fabric-private-chaincode/extension/fsc/integration/nwo/fabric/fpc/externalbuilders"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var logger = logging.MustGetLogger("integration.nwo.fabric.fpc")

// hasFPCEnabled reports whether any FPC chaincode is registered on the topology.
// Replaces the removed topology.Topology.FPC boolean field (dropped in FSC PR #696).
func hasFPCEnabled(t *topology.Topology) bool {
	for _, cc := range t.Chaincodes {
		if cc.Private {
			return true
		}
	}
	return false
}

type ChaincodeInput struct {
	Args []string `json:"Args"`
}

type Connection struct {
	Address     string `json:"address"`
	DialTimeout string `json:"dial_timeout"`
	TlsRequired bool   `json:"tls_required"`
}

type Extension struct {
	network *nnetwork.Network

	ports   map[string][]uint16
	FPCERCC *topology.ChannelChaincode
}

func NewExtension(network *nnetwork.Network) *Extension {
	return &Extension{
		network: network,
		ports:   map[string][]uint16{},
	}
}

func (n *Extension) CheckTopology() {
	if !hasFPCEnabled(n.network.Topology()) {
		return
	}

	path := n.prepareExternalBuilderScripts()
	n.network.ExternalBuilders = append(n.network.ExternalBuilders, fabricconfig.ExternalBuilder{
		Path:                 path,
		Name:                 "fpc-external-launcher",
		PropagateEnvironment: []string{"CORE_PEER_ID", "FABRIC_LOGGING_SPEC"},
	})

	// Reserve ports for each private chaincode
	for _, chaincode := range n.network.Topology().Chaincodes {
		if !chaincode.Private {
			continue
		}

		if chaincode.Chaincode.Name == "ercc" {
			n.FPCERCC = chaincode
		}

		var ports []uint16
		for range chaincode.Peers {
			ports = append(ports, n.network.Context.ReservePort())
		}

		n.ports[chaincode.Chaincode.Name] = ports
	}
}

func (n *Extension) GenerateArtifacts() {
	for _, chaincode := range n.network.Topology().Chaincodes {
		if !chaincode.Private {
			continue
		}
		n.preparePackage(chaincode)
	}

	for _, org := range n.network.PeerOrgs() {
		var p *topology.Peer
		for _, peer := range n.network.Peers {
			if peer.Organization == org.Name {
				p = peer
				break
			}
		}
		Expect(p).NotTo(BeNil())

	}
}

func (n *Extension) PostRun(load bool) {
	if !hasFPCEnabled(n.network.Topology()) {
		return
	}

	if load {
		panic("Load is not supported. Restart from scratch.")
	}

	// deploy and run containers
	if len(n.network.Topology().Chaincodes) != 0 {
		for _, chaincode := range n.network.Topology().Chaincodes {
			if chaincode.Private {
				n.reservePorts(chaincode)
				n.preparePackage(chaincode)
				n.deployChaincode(chaincode)
			}
		}
	}

	// generate connection profiles
	for _, org := range n.network.PeerOrgs() {
		err := n.generateConnections(org.Name)
		Expect(err).ToNot(HaveOccurred())
	}

	// Give some time to the enclave to get up and running...
	time.Sleep(3 * time.Second)
}

func (n *Extension) deployChaincode(chaincode *topology.ChannelChaincode) {
	if !hasFPCEnabled(n.network.Topology()) {
		return
	}

	orderer := n.network.Orderer("orderer")
	peers := n.network.PeersByName(chaincode.Peers)
	// Install on each peer its own chaincode package
	var packageIDs []string
	for _, peer := range peers {
		chaincode.Chaincode.PackageID = ""
		org := n.network.Organization(peer.Organization)
		chaincode.Chaincode.PackageFile = filepath.Join(
			n.packagePath(chaincode.Chaincode.Name),
			fmt.Sprintf("%s.%s.%s.tgz", chaincode.Chaincode.Name, peer.Name, org.Domain),
		)
		nnetwork.InstallChaincode(n.network, &chaincode.Chaincode, peer)
		nnetwork.ApproveChaincodeForMyOrg(n.network, chaincode.Channel, orderer, &chaincode.Chaincode, peer)
		packageIDs = append(packageIDs, chaincode.Chaincode.PackageID)
	}

	n.runDockerContainers(chaincode, packageIDs)

	nnetwork.CheckCommitReadinessUntilReady(n.network, chaincode.Channel, &chaincode.Chaincode, n.network.PeerOrgsByPeers(peers), peers...)
	nnetwork.CommitChaincode(n.network, chaincode.Channel, orderer, &chaincode.Chaincode, peers[0], peers...)
	time.Sleep(10 * time.Second)
	for i, peer := range peers {
		nnetwork.QueryInstalledReferences(n.network,
			chaincode.Channel, chaincode.Chaincode.Label, packageIDs[i],
			peer,
			[]string{chaincode.Chaincode.Name, chaincode.Chaincode.Version},
		)
	}
	if chaincode.Chaincode.InitRequired {
		nnetwork.InitChaincode(n.network, chaincode.Channel, orderer, &chaincode.Chaincode, peers...)
	}

	// initEnclave if not an ERCC
	if chaincode.Chaincode.Name != "ercc" {
		n.initEnclave(chaincode)
	}
}

func (n *Extension) packagePath(suffix string) string {
	return filepath.Join(n.network.Context.RootDir(), n.network.Prefix, "fpc", suffix)
}

func (n *Extension) runDockerContainers(chaincode *topology.ChannelChaincode, packageIDs []string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	Expect(err).ToNot(HaveOccurred())

	var mounts []mount.Mount
	var devices []container.DeviceMapping

	sgxMode := strings.ToUpper(chaincode.PrivateChaincode.SGXMode)
	err = validateSGXMode(sgxMode)
	Expect(err).ToNot(HaveOccurred())

	if sgxMode == SGX_MODE_HW {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: "/var/run/aesmd",
			Target: "/var/run/aesmd",
		})

		err := validateSGXDevicesPaths(chaincode.PrivateChaincode.SGXDevicesPaths)
		Expect(err).ToNot(HaveOccurred())

		// add sgx devices to the container depending on the driver installed on the platform
		// we may attach /dev/isgx, or /dev/sgx/enclave and /dev/sgx/provision
		for _, devicePath := range chaincode.PrivateChaincode.SGXDevicesPaths {
			devices = append(devices, container.DeviceMapping{
				PathOnHost:        devicePath,
				PathInContainer:   devicePath,
				CgroupPermissions: "rwm",
			})
		}
	}

	// attach dcap qpl if exists
	qcnlConfigPath := "/etc/sgx_default_qcnl.conf"
	if exists, err := pathExists(qcnlConfigPath); exists && err == nil {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: qcnlConfigPath,
			Target: qcnlConfigPath,
		})
	}

	peers := n.network.PeersByName(chaincode.Peers)
	for i, peer := range peers {
		port := strconv.Itoa(int(n.ports[chaincode.Chaincode.Name][i]))
		containerName := fmt.Sprintf("%s.%s.%s.%s",
			n.network.NetworkID,
			chaincode.Chaincode.Name, peer.Name,
			n.network.Organization(peer.Organization).Domain)

		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: chaincode.PrivateChaincode.Image,
			Tty:   false,
			Env: []string{
				"CHAINCODE_SERVER_ADDRESS=0.0.0.0:" + port,
				"CHAINCODE_PKG_ID=" + packageIDs[i],
				"FABRIC_LOGGING_SPEC=debug",
				fmt.Sprintf("SGX_MODE=%s", sgxMode),
			},
			ExposedPorts: nat.PortSet{
				nat.Port(port + "/tcp"): struct{}{},
			},
		}, &container.HostConfig{
			Mounts: mounts,
			Resources: container.Resources{
				Devices: devices,
			},
			PortBindings: nat.PortMap{
				nat.Port(port + "/tcp"): []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: port,
					},
				},
			},
		},
			nil, nil, containerName,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(cli.ContainerStart(ctx, resp.ID, container.StartOptions{})).ToNot(HaveOccurred())
		time.Sleep(3 * time.Second)

		dockerLogger := logging.MustGetLogger("fpc.container." + containerName)
		go func() {
			reader, err := cli.ContainerLogs(context.Background(), resp.ID, container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Follow:     true,
				Timestamps: false,
			})
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				dockerLogger.Debugf("%s", scanner.Text())
			}
		}()
	}
}

const (
	SGX_MODE_HW  = "HW"
	SGX_MODE_SIM = "SIM"
)

func validateSGXMode(mode string) error {
	if strings.ToUpper(mode) == SGX_MODE_HW || strings.ToUpper(mode) == SGX_MODE_SIM {
		return nil
	}
	return errors.Errorf("invalid SGX Mode = %s", mode)
}

func validateSGXDevicesPaths(paths []string) error {
	// remove duplicates
	devicesPaths := make(map[string]bool, len(paths))
	for _, p := range paths {
		devicesPaths[p] = true
	}

	// check exists
	for p := range devicesPaths {
		if exists, err := pathExists(p); !exists || err != nil {
			return errors.Wrapf(err, "invalid SGX device path = %s", p)
		}
	}

	// check some extra rules
	allowedSettings := [][]string{
		{"/dev/sgx/isgx"},
		{"/dev/sgx/enclave"},
		{"/dev/sgx/enclave", "/dev/sgx/provision"},
	}

	for _, setting := range allowedSettings {
		// TODO implement me
		_ = setting
	}

	return nil
}

func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "error checking path = %s", path)
	}
}

func (n *Extension) initEnclave(chaincode *topology.ChannelChaincode) {
	initOrgs := map[string]bool{}
	var erccPeerAddresses []string
	peers := n.network.PeersByName(n.FPCERCC.Peers)
	for _, p := range peers {
		if exists := initOrgs[p.Organization]; !exists {
			erccPeerAddresses = append(erccPeerAddresses, n.network.PeerAddress(p, nnetwork.ListenPort))
			initOrgs[p.Organization] = true
		}
	}

	orderer := n.network.Orderer("orderer")
	peers = n.network.PeersByName(chaincode.Peers)
	attestationParams := CreateSIMAttestationParams()
	for _, peer := range peers {
		lifecycleClient, err := lifecycle.New(func(channelID string) (lifecycle.ChannelClient, error) {
			return &ChannelClient{
				n:         n,
				peer:      peer,
				orderer:   orderer,
				chaincode: chaincode,
			}, nil
		})
		Expect(err).ToNot(HaveOccurred())
		_, err = lifecycleClient.LifecycleInitEnclave(chaincode.Channel, lifecycle.LifecycleInitEnclaveRequest{
			ChaincodeID:         chaincode.Chaincode.Name,
			EnclavePeerEndpoint: erccPeerAddresses[0],
			AttestationParams:   attestationParams,
		})
		Expect(err).ToNot(HaveOccurred())

		// Enclave Init should happen once only
		return
	}
}

func (n *Extension) preparePackage(chaincode *topology.ChannelChaincode) {
	// Generate Enclave Registry CC package for each peer that will conn
	Expect(os.MkdirAll(n.packagePath(chaincode.Chaincode.Name), 0770)).NotTo(HaveOccurred())

	peers := n.network.PeersByName(chaincode.Peers)
	for i, peer := range peers {
		org := n.network.Organization(peer.Organization)

		packageFilePath := filepath.Join(n.packagePath(chaincode.Chaincode.Name), fmt.Sprintf("%s.%s.%s.tgz", chaincode.Chaincode.Name, peer.Name, org.Domain))
		if _, err := os.Stat(packageFilePath); os.IsNotExist(err) {
			Expect(packager.New().PackageChaincode(
				chaincode.Chaincode.Name,
				chaincode.Chaincode.Lang,
				chaincode.Chaincode.Label,
				packageFilePath,
				func(s string, s2 string) (string, []byte) {
					if strings.HasSuffix(s, "connection.json") {
						raw, err := json.MarshalIndent(&Connection{
							Address:     "127.0.0.1:" + strconv.Itoa(int(n.ports[chaincode.Chaincode.Name][i])),
							DialTimeout: "10s",
							TlsRequired: false,
						}, "", " ")
						Expect(err).NotTo(HaveOccurred())
						return filepath.Join(fmt.Sprintf("%s.%s", peer.Name, org.Domain), "connection.json"), raw
					}
					return "", nil
				},
			)).ToNot(HaveOccurred())
		}
	}
}

func (n *Extension) reservePorts(chaincode *topology.ChannelChaincode) {
	if ports, ok := n.ports[chaincode.Chaincode.Name]; ok && len(ports) != 0 {
		return
	}

	var ports []uint16
	for range chaincode.Peers {
		ports = append(ports, n.network.Context.ReservePort())
	}
	n.ports[chaincode.Chaincode.Name] = ports
}

func (n *Extension) externalBuilderPath() string {
	return filepath.Join(n.network.Context.RootDir(), n.network.Prefix, "fpc", "externalbuilders", "chaincode_server")
}

func (n *Extension) prepareExternalBuilderScripts() string {
	path := n.externalBuilderPath()
	Expect(os.MkdirAll(filepath.Join(path, "bin"), 0777)).NotTo(HaveOccurred())

	// Store External Builder scripts to file
	for _, script := range []struct{ name, content string }{
		{"build", externalbuilders.Build},
		{"detect", externalbuilders.Detect},
		{"release", externalbuilders.Release},
	} {
		scriptPath := filepath.Join(path, "bin", script.name)
		Expect(os.WriteFile(scriptPath, []byte(script.content), 0777)).NotTo(HaveOccurred())
	}

	return path
}
