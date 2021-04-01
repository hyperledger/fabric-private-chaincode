/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt"
	fpc "github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("sdk-test")

var testNetworkPath string

func firstFileInPath(dir string) (string, error) {
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}
	if len(files) != 1 {
		return "", fmt.Errorf("folder should contain only a single file")
	}
	return filepath.Join(dir, files[0].Name()), nil
}

func populateWallet(wallet *gateway.Wallet) error {
	logger.Debugf("============ Populating wallet ============")
	credPath := filepath.Join(
		testNetworkPath,
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath, err := firstFileInPath(filepath.Join(credPath, "signcerts"))
	if err != nil {
		return err
	}

	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyPath, err := firstFileInPath(filepath.Join(credPath, "keystore"))
	if err != nil {
		return err
	}

	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func main() {
	withLifecycleInitEnclave := flag.Bool("withLifecycleInitEnclave", false, "run with lifecycleInitEnclave")
	flag.Parse()

	os.Setenv("GRPC_TRACE", "all")
	os.Setenv("GRPC_VERBOSITY", "DEBUG")
	os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "INFO")

	ccID := os.Getenv("CC_ID")
	if ccID == "" {
		panic("CC_ID not set")
	}

	fpcPath := os.Getenv("FPC_PATH")
	if fpcPath == "" {
		panic("FPC_PATH not set")
	}
	testNetworkPath = filepath.Join(fpcPath, "integration", "test-network", "fabric-samples", "test-network")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		logger.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		logger.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			logger.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		testNetworkPath,
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		logger.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	channelID := "mychannel"

	network, err := gw.GetNetwork(channelID)
	if err != nil {
		logger.Fatalf("Failed to get network: %v", err)
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// FPC Client SDK Lifecycle API example
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// if we call this app with the "-withLifecycleInitEnclave" flag
	if *withLifecycleInitEnclave {

		fmt.Println("---------- SDK -------------")

		sdk, err := fabsdk.New(config.FromFile(filepath.Clean(ccpPath)))
		if err != nil {
			logger.Fatalf("failed to create sdk: %v", err)
		}
		defer sdk.Close()

		orgAdmin := "Admin"
		orgName := "org1"

		adminContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
		adminClient, err := resmgmt.New(adminContext)
		if err != nil {
			logger.Fatal(err)
		}

		attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
		if err != nil {
			logger.Fatal(err)
		}

		initReq := resmgmt.LifecycleInitEnclaveRequest{
			ChaincodeID:         ccID,
			EnclavePeerEndpoint: "peer0.org1.example.com", // define the peer where we wanna init our enclave
			AttestationParams:   attestationParams,
		}

		logger.Infof("--> Invoke LifecycleInitEnclave")
		_, err = adminClient.LifecycleInitEnclave(channelID, initReq)
		if err != nil {
			logger.Fatalf("Failed to invoke LifecycleInitEnclave: %v", err)
		}
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// FPC Client SDK contract API example
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Get FPC Contract
	contract := fpc.GetContract(network, ccID)

	// Invoke FPC Chaincode
	logger.Infof("--> Invoke FPC chaincode: ")
	result, err := contract.SubmitTransaction("myFunction", "arg1", "arg2", "arg3")
	if err != nil {
		logger.Fatalf("Failed to Submit transaction: %v", err)
	}
	logger.Infof("--> Result: %s", string(result))
}
