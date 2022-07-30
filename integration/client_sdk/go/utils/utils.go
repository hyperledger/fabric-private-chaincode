/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	fpcmgmt "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt"
	fpcpackager "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/fab/ccpackager"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	lifecyclepkg "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric/common/flogging"
)

var (
	FPCPath         string
	testNetworkPath string
	ccpPath         string
	logger          = flogging.MustGetLogger("client_sdk_test")
)

type chaincodeDetails struct {
	Id           string
	Path         string
	Version      string
	Lang         string
	Policy       *common.SignaturePolicyEnvelope
	Seq          int64
	Vscc         string
	Escc         string
	InitRequired bool
	Package      []byte
	PackageID    string
}

type networkDetails struct {
	ChannelID    string
	Peers        []string
	Orderers     []string
	EnclavePeers []string
	CaaSEndpoint map[string]string
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

	ccpPath = filepath.Join(
		testNetworkPath,
		"config",
		"msp",
		"connection.yaml",
	)
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		testNetworkPath,
		"config",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "peer.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("SampleOrg", string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func SetupNetwork(channel string) (*gateway.Network, error) {

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "false")
	if err != nil {
		return nil, fmt.Errorf("error setting DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			return nil, fmt.Errorf("failed to populate wallet contents: %v", err)
		}
	}

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %v", err)
	}

	return network, nil
}

func Setup(ccID, ccPath string, initEnclave bool) error {

	var enclavePeers []string
	if initEnclave {
		enclavePeers = append(enclavePeers, "jdoe_test.sampleorg.example.com")
	}

	nwDetails := &networkDetails{
		ChannelID:    "mychannel",
		Peers:        []string{"jdoe_test.sampleorg.example.com"},
		Orderers:     []string{"orderer.example.com"},
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

	// get sdk instance
	sdk, err := fabsdk.New(config.FromFile(filepath.Clean(ccpPath)))
	if err != nil {
		return fmt.Errorf("failed to create sdk: %v", err)
	}
	defer sdk.Close()

	// new client
	orgAdmin := "Admin"
	orgName := "SampleOrg"

	adminContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))

	client, err := fpcmgmt.New(adminContext)
	if err != nil {
		return fmt.Errorf("failed to create context: %v", err)
	}

	// install fpc chaincode
	err = installChaincode(client, ccDetails, nwDetails)
	if err != nil {
		return fmt.Errorf("error during installing chaincode: %v", err)
	}

	return nil
}

func installChaincode(client *fpcmgmt.Client, cc *chaincodeDetails, nw *networkDetails) error {
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
		return fmt.Errorf("failed to create new chaincode package: %v", err)
	}
	logger.Infof("%s successfully packaged", cc.Id)

	// install
	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   cc.Id,
		Package: ccPkg,
	}
	resp, err := client.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return fmt.Errorf("failed to install chaincode: %v", err)
	}
	packageID := lifecyclepkg.ComputePackageID(installCCReq.Label, installCCReq.Package)
	logger.Infof("%s successfully installed: %v", cc.Id, resp)

	// approveformyorg
	approveCCReq := resmgmt.LifecycleApproveCCRequest{
		Name:              cc.Id,
		Version:           cc.Version,
		PackageID:         packageID,
		Sequence:          cc.Seq,
		EndorsementPlugin: cc.Escc,
		ValidationPlugin:  cc.Vscc,
		SignaturePolicy:   cc.Policy,
		InitRequired:      cc.InitRequired,
	}
	txid, err := client.LifecycleApproveCC(nw.ChannelID, approveCCReq,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(nw.Peers...),
		resmgmt.WithOrdererEndpoint(nw.Orderers[0]),
	)
	if err != nil {
		return fmt.Errorf("failed to approve chaincode: %v", err)
	}
	logger.Infof("%s successfully approved with txid: %v", cc.Id, txid)

	// checkcommitreadiness
	checkCCReq := resmgmt.LifecycleCheckCCCommitReadinessRequest{
		Name:              cc.Id,
		Version:           cc.Version,
		Sequence:          cc.Seq,
		EndorsementPlugin: cc.Escc,
		ValidationPlugin:  cc.Vscc,
		SignaturePolicy:   cc.Policy,
		InitRequired:      cc.InitRequired,
	}
	chkresp, err := client.LifecycleCheckCCCommitReadiness(nw.ChannelID, checkCCReq,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(nw.Peers...),
	)
	if err != nil {
		return fmt.Errorf("failed to check chaincode commit readiness: %v", err)
	}
	logger.Infof("%s readiness check: %v", cc.Id, chkresp)

	// commit
	commitCCReq := resmgmt.LifecycleCommitCCRequest{
		Name:              cc.Id,
		Version:           cc.Version,
		Sequence:          cc.Seq,
		EndorsementPlugin: cc.Escc,
		ValidationPlugin:  cc.Vscc,
		SignaturePolicy:   cc.Policy,
		InitRequired:      cc.InitRequired,
	}
	_, err = client.LifecycleCommitCC(nw.ChannelID, commitCCReq,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(nw.Peers...),
		resmgmt.WithOrdererEndpoint(nw.Orderers[0]),
	)
	if err != nil {
		return fmt.Errorf("failed to commit chaincode: %v", err)
	}
	logger.Infof("%s successfully committed", cc.Id)

	// init enclave
	if len(nw.EnclavePeers) > 0 {

		attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
		if err != nil {
			return fmt.Errorf("failed to load attestation params from environment: %v", err)
		}

		initReq := fpcmgmt.LifecycleInitEnclaveRequest{
			ChaincodeID:         cc.Id,
			EnclavePeerEndpoint: nw.EnclavePeers[0], // define the peer where we wanna init our enclave
			AttestationParams:   attestationParams,
		}

		initTxId, err := client.LifecycleInitEnclave(nw.ChannelID, initReq,
			// Note that these options are currently ignored by our implementation
			resmgmt.WithRetry(retry.DefaultResMgmtOpts),
			resmgmt.WithTargetEndpoints(nw.Peers...), // peers that are responsible for enclave registration
			resmgmt.WithOrdererEndpoint(nw.Orderers[0]),
		)
		if err != nil {
			return fmt.Errorf("failed to init enclave: %v", err)
		}
		logger.Infof("%s successfully initialized enclave: %v", cc.Id, initTxId)
	} else {
		logger.Infof("%s Skip enclave initialization", cc.Id)
	}

	return nil
}
