/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-lib-go/common/flogging"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

var logger = flogging.MustGetLogger("sdk-test")

// connectionProfile contains the subset of a Fabric connection profile needed to reach a peer
type connectionProfile struct {
	Peers map[string]struct {
		URL        string `yaml:"url"`
		TLSCACerts struct {
			Pem string `yaml:"pem"`
		} `yaml:"tlsCACerts"`
		GRPCOptions struct {
			SSLTargetNameOverride string `yaml:"ssl-target-name-override"`
		} `yaml:"grpcOptions"`
	} `yaml:"peers"`
}

// newPeerConnection creates a gRPC connection to the given peer as defined in the connection profile
func newPeerConnection(ccpPath, peerName string) (*grpc.ClientConn, error) {
	data, err := os.ReadFile(filepath.Clean(ccpPath))
	if err != nil {
		return nil, err
	}
	ccp := &connectionProfile{}
	if err := yaml.Unmarshal(data, ccp); err != nil {
		return nil, err
	}
	peer, ok := ccp.Peers[peerName]
	if !ok {
		return nil, fmt.Errorf("peer %s not found in %s", peerName, ccpPath)
	}

	endpoint := peer.URL
	if i := strings.Index(endpoint, "://"); i >= 0 {
		endpoint = endpoint[i+3:]
	}
	serverName := peer.GRPCOptions.SSLTargetNameOverride
	if serverName == "" {
		serverName = peerName
	}

	return fpc.NewGrpcConnection(endpoint, []byte(peer.TLSCACerts.Pem), serverName)
}

// connectGateway connects to the Fabric Gateway using the identity found in the given MSP folder
func connectGateway(conn *grpc.ClientConn, mspId, mspPath string) (*client.Gateway, identity.Sign, error) {
	id, sign, err := fpc.NewIdentityFromMSP(mspId, mspPath)
	if err != nil {
		return nil, nil, err
	}
	gw, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(conn))
	return gw, sign, err
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

	// let's make sure FPC_PATH is set
	fpcPath := getEnvWithPanic("FPC_PATH")

	ccID := getEnvWithFallback("CC_ID", "echo")
	channelID := getEnvWithFallback("CHANNEL_ID", "mychannel")
	orgName := getEnvWithFallback("ORG_NAME", "Org1")
	orgNameFull := getEnvWithFallback("ORG_NAME_FULL", strings.ToLower(orgName)+".example.com")
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

	// If not set we use the admin msp folder in the fabric-samples test network
	adminMspPath := getEnvWithFallback("ADMIN_MSP_PATH", filepath.Join(
		testNetworkPath,
		"organizations",
		"peerOrganizations",
		orgNameFull,
		"users",
		fmt.Sprintf("Admin@%s", orgNameFull),
		"msp",
	))

	// the peer that hosts the enclave and that we use as gateway
	peerName := getEnvWithFallback("PEER_NAME", "peer0."+orgNameFull)

	// If not set we use the connection profile in the fabric-samples test network
	ccpPath := getEnvWithFallback("CCP_PATH", filepath.Join(
		testNetworkPath,
		"organizations",
		"peerOrganizations",
		orgNameFull,
		fmt.Sprintf("connection-%s.yaml", strings.ToLower(orgName)),
	))

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// FPC Client SDK Lifecycle API example
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	conn, err := newPeerConnection(ccpPath, peerName)
	if err != nil {
		logger.Fatalf("Failed to connect to peer: %v", err)
	}
	defer conn.Close()

	// enclave peers are called directly; their endpoints, as registered at the enclave registry,
	// are the peer names of the connection profile
	dialPeer := func(endpoint string) (grpc.ClientConnInterface, error) {
		return newPeerConnection(ccpPath, endpoint)
	}

	// if we call this app with the "-withLifecycleInitEnclave" flag
	if *withLifecycleInitEnclave {

		adminGw, adminSign, err := connectGateway(conn, mspId, adminMspPath)
		if err != nil {
			logger.Fatalf("Failed to connect to gateway as admin: %v", err)
		}
		defer adminGw.Close()

		adminClient, err := resmgmt.New(adminGw, adminSign, dialPeer)
		if err != nil {
			logger.Fatal(err)
		}

		attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
		if err != nil {
			logger.Fatal(err)
		}

		initReq := resmgmt.LifecycleInitEnclaveRequest{
			ChaincodeID:         ccID,
			EnclavePeerEndpoint: peerName, // define the peer where we wanna init our enclave
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

	gw, sign, err := connectGateway(conn, mspId, mspPath)
	if err != nil {
		logger.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network := gw.GetNetwork(channelID)

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// FPC Client SDK contract API example
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Get FPC Contract
	contract := fpc.GetContract(network, ccID, sign, dialPeer)

	// Invoke FPC Chaincode
	logger.Infof("--> Invoke FPC chaincode: %s", contract.Name())
	result, err := contract.SubmitTransaction("myFunction", "arg1", "arg2", "arg3")
	if err != nil {
		logger.Fatalf("Failed to Submit transaction: %v", err)
	}
	logger.Infof("--> Result: %s", string(result))
}
