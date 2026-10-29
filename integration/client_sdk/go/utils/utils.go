/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	adminidentity "github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-lib-go/common/flogging"
	fpcmgmt "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt"
	fpcpackager "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/fab/ccpackager"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
)

const (
	mspID        = "SampleOrg"
	peerEndpoint = "127.0.0.1:7051"
	// enclavePeer is the endpoint under which the enclave of the test network peer is registered
	enclavePeer = "jdoe_test.sampleorg.example.com"
)

var (
	FPCPath         string
	testNetworkPath string
	mspPath         string
	logger          = flogging.MustGetLogger("client_sdk_test")
)

type chaincodeDetails struct {
	Id           string
	Path         string
	Version      string
	Lang         string
	Policy       *peer.ApplicationPolicy
	Seq          int64
	Vscc         string
	Escc         string
	InitRequired bool
	Package      []byte
	PackageID    string
}

type networkDetails struct {
	ChannelID    string
	EnclavePeers []string
}

func init() {
	os.Setenv("GRPC_TRACE", "all")
	os.Setenv("GRPC_VERBOSITY", "DEBUG")
	os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "INFO")

	FPCPath = os.Getenv("FPC_PATH")
	if FPCPath == "" {
		panic("FPC_PATH not set")
	}

	testNetworkPath = filepath.Join(FPCPath, "integration")

	mspPath = filepath.Join(
		testNetworkPath,
		"config",
		"msp",
	)
}

// connect creates a gRPC connection to the Fabric Gateway of the test network peer
func connect() (*grpc.ClientConn, error) {
	conn, err := fpc.NewGrpcConnection(peerEndpoint, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}
	return conn, nil
}

// dialPeer connects to an enclave peer; the test network has a single peer, registered under enclavePeer
func dialPeer(endpoint string) (grpc.ClientConnInterface, error) {
	if endpoint != enclavePeer {
		return nil, fmt.Errorf("unknown peer %s", endpoint)
	}
	return connect()
}

func newGateway(conn *grpc.ClientConn) (*client.Gateway, identity.Sign, error) {
	id, sign, err := fpc.NewIdentityFromMSP(mspID, mspPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load identity: %w", err)
	}

	gw, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(conn))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}
	return gw, sign, nil
}

// Network is a Fabric Gateway network together with the signing function of its identity
type Network struct {
	*client.Network
	sign identity.Sign
}

func SetupNetwork(channel string) (*Network, error) {
	conn, err := connect()
	if err != nil {
		return nil, err
	}

	gw, sign, err := newGateway(conn)
	if err != nil {
		return nil, err
	}

	return &Network{Network: gw.GetNetwork(channel), sign: sign}, nil
}

// GetContract returns the FPC contract ccID, calling the enclave peer directly
func GetContract(network *Network, ccID string) fpc.Contract {
	return fpc.GetContract(network.Network, ccID, network.sign, dialPeer)
}

func Setup(ccID, ccPath string, initEnclave bool) error {

	var enclavePeers []string
	if initEnclave {
		enclavePeers = append(enclavePeers, enclavePeer)
	}

	nwDetails := &networkDetails{
		ChannelID:    "mychannel",
		EnclavePeers: enclavePeers,
	}

	mrenclave, err := fpcpackager.ReadMrenclave(ccPath)
	if err != nil {
		return err
	}

	ccDetails := &chaincodeDetails{
		Id:           ccID,
		Path:         ccPath,
		Version:      mrenclave,
		Lang:         fpcpackager.ChaincodeType,
		Seq:          int64(1),
		Vscc:         "vscc",
		Escc:         "escc",
		InitRequired: false,
	}

	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// install fpc chaincode
	err = installChaincode(conn, ccDetails, nwDetails)
	if err != nil {
		return fmt.Errorf("error during installing chaincode: %w", err)
	}

	return nil
}

func installChaincode(conn *grpc.ClientConn, cc *chaincodeDetails, nw *networkDetails) error {
	// TODO promote this install function to a test suite for the FPC Admin API
	// Right now this is used by the auction test to setup the "environment"; additional testing
	// of the FPC Admin API would be good here.

	// get sgx mode
	sgxMode := os.Getenv(sgx.SGXModeEnvKey)
	if sgxMode == "" {
		return errors.New("sgx mode is not set via env vars")
	}

	// package
	desc := &fpcpackager.Descriptor{
		Path:    cc.Path,
		Type:    cc.Lang,
		Label:   cc.Id,
		SGXMode: sgxMode,
	}
	ccPkg, err := fpcpackager.NewCCPackage(desc)
	if err != nil {
		return fmt.Errorf("failed to create new chaincode package: %w", err)
	}
	logger.Infof("%s successfully packaged", cc.Id)

	certificate, err := adminidentity.ReadCertificate(filepath.Join(mspPath, "signcerts", "peer.pem"))
	if err != nil {
		return fmt.Errorf("failed to read admin certificate: %w", err)
	}
	privateKey, err := adminidentity.ReadPrivateKey(filepath.Join(mspPath, "keystore", "key.pem"))
	if err != nil {
		return fmt.Errorf("failed to read admin private key: %w", err)
	}
	admin, err := adminidentity.NewPrivateKeySigningIdentity(mspID, certificate, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create admin identity: %w", err)
	}

	ctx := context.Background()

	// install
	resp, err := chaincode.NewPeer(conn, admin).Install(ctx, bytes.NewReader(ccPkg))
	if err != nil {
		return fmt.Errorf("failed to install chaincode: %w", err)
	}
	packageID := chaincode.GetPackageID(cc.Id, ccPkg)
	logger.Infof("%s successfully installed: %v", cc.Id, resp)

	definition := &chaincode.Definition{
		ChannelName:       nw.ChannelID,
		PackageID:         packageID,
		Name:              cc.Id,
		Version:           cc.Version,
		EndorsementPlugin: cc.Escc,
		ValidationPlugin:  cc.Vscc,
		Sequence:          cc.Seq,
		ApplicationPolicy: cc.Policy,
		InitRequired:      cc.InitRequired,
	}
	lifecycleGateway := chaincode.NewGateway(conn, admin)

	// approveformyorg
	err = lifecycleGateway.Approve(ctx, definition)
	if err != nil {
		return fmt.Errorf("failed to approve chaincode: %w", err)
	}
	logger.Infof("%s successfully approved", cc.Id)

	// checkcommitreadiness
	chkresp, err := lifecycleGateway.CheckCommitReadiness(ctx, definition)
	if err != nil {
		return fmt.Errorf("failed to check chaincode commit readiness: %w", err)
	}
	logger.Infof("%s readiness check: %v", cc.Id, chkresp)

	// commit
	err = lifecycleGateway.Commit(ctx, definition)
	if err != nil {
		return fmt.Errorf("failed to commit chaincode: %w", err)
	}
	logger.Infof("%s successfully committed", cc.Id)

	// init enclave
	if len(nw.EnclavePeers) > 0 {

		attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
		if err != nil {
			return fmt.Errorf("failed to load attestation params from environment: %w", err)
		}

		initReq := fpcmgmt.LifecycleInitEnclaveRequest{
			ChaincodeID:         cc.Id,
			EnclavePeerEndpoint: nw.EnclavePeers[0], // define the peer where we wanna init our enclave
			AttestationParams:   attestationParams,
		}

		gw, sign, err := newGateway(conn)
		if err != nil {
			return err
		}
		defer gw.Close()

		client, err := fpcmgmt.New(gw, sign, dialPeer)
		if err != nil {
			return fmt.Errorf("failed to create FPC management client: %w", err)
		}

		initTxId, err := client.LifecycleInitEnclave(nw.ChannelID, initReq)
		if err != nil {
			return fmt.Errorf("failed to init enclave: %w", err)
		}
		logger.Infof("%s successfully initialized enclave: %v", cc.Id, initTxId)
	} else {
		logger.Infof("%s Skip enclave initialization", cc.Id)
	}

	return nil
}
