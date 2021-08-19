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

	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("sdk-test")

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

func populateWallet(wallet *gateway.Wallet, mspPath, mspId, userId string) error {
	certPath, err := firstFileInPath(filepath.Join(mspPath, "signcerts"))
	if err != nil {
		return err
	}

	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyPath, err := firstFileInPath(filepath.Join(mspPath, "keystore"))
	if err != nil {
		return err
	}

	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity(mspId, string(cert), string(key))

	return wallet.Put(userId, identity)
}

func getEnvWithFallback(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func getEnvWithPanic(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic(key + " not set")
	}
	return value
}

func setEnvWithPanic(key, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		panic(fmt.Sprintf("Error setting %s environment variable: %v", key, err))
	}
}

func main() {
	withLifecycleInitEnclave := flag.Bool("withLifecycleInitEnclave", false, "run with lifecycleInitEnclave")
	flag.Parse()

	setEnvWithPanic("GRPC_TRACE", "all")
	setEnvWithPanic("GRPC_VERBOSITY", "DEBUG")
	setEnvWithPanic("GRPC_GO_LOG_SEVERITY_LEVEL", "INFO")
	setEnvWithPanic("DISCOVERY_AS_LOCALHOST", "true")

	// let's make sure FPC_PATH is set
	fpcPath := getEnvWithPanic("FPC_PATH")

	ccID := getEnvWithFallback("CC_ID", "echo")
	channelID := getEnvWithFallback("CHANNEL_ID", "mychannel")
	orgName := getEnvWithFallback("ORG_NAME", "org1")
	orgNameFull := getEnvWithFallback("ORG_NAME_FULL", orgName+".example.com")
	mspId := getEnvWithFallback("MSP_ID", orgName+"MSP")
	userId := getEnvWithFallback("USER_ID", "User1")

	// we use the mspPath and connections profil provided by the fabric-samples test network if not specified by the user
	testNetworkPath := filepath.Join(fpcPath, "samples", "deployment", "test-network", "fabric-samples", "test-network")

	// If not set we use the msp folder in the fabric-samples test network
	mspPath := getEnvWithFallback("MSP_PATH", filepath.Join(
		testNetworkPath,
		"organizations",
		"peerOrganizations",
		orgNameFull,
		"users",
		fmt.Sprintf("%s@%s", userId, orgNameFull),
		"msp",
	))

	// If not set we use the msp folder in the fabric-samples test network
	ccpPath := getEnvWithFallback("CCP_PATH", filepath.Join(
		testNetworkPath,
		"organizations",
		"peerOrganizations",
		orgNameFull,
		fmt.Sprintf("connection-%s.yaml", orgName),
	))

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// FPC Client SDK Lifecycle API example
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// if we call this app with the "-withLifecycleInitEnclave" flag
	if *withLifecycleInitEnclave {

		sdk, err := fabsdk.New(config.FromFile(filepath.Clean(ccpPath)))
		if err != nil {
			logger.Fatalf("failed to create sdk: %v", err)
		}
		defer sdk.Close()

		orgAdmin := "Admin"

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
			EnclavePeerEndpoint: "peer0." + orgNameFull, // define the peer where we wanna init our enclave
			AttestationParams:   attestationParams,
		}

		logger.Infof("--> Invoke LifecycleInitEnclave")
		_, err = adminClient.LifecycleInitEnclave(channelID, initReq)
		if err != nil {
			logger.Fatalf("Failed to invoke LifecycleInitEnclave: %v", err)
		}
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Setup Fabric Gateway
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	wallet := gateway.NewInMemoryWallet()
	err := populateWallet(wallet, mspPath, mspId, userId)
	if err != nil {
		logger.Fatalf("Failed to populate wallet contents: %v", err)
	}

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, userId),
	)
	if err != nil {
		logger.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channelID)
	if err != nil {
		logger.Fatalf("Failed to get network: %v", err)
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// FPC Client SDK contract API example
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Get FPC Contract
	contract := fpc.GetContract(network, ccID)

	// Invoke FPC Chaincode
	logger.Infof("--> Invoke FPC chaincode: %s", contract.Name())
	result, err := contract.SubmitTransaction("myFunction", "arg1", "arg2", "arg3")
	if err != nil {
		logger.Fatalf("Failed to Submit transaction: %v", err)
	}
	logger.Infof("--> Result: %s", string(result))
}
