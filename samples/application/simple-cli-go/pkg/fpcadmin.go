/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkg

import (
	"fmt"
	"path/filepath"

	cfg "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/fgosdkresmgmt"
	fpcmgmt "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
)

type Admin struct {
	sdk         *fabsdk.FabricSDK
	client      *fgosdkresmgmt.Client
	config      *Config
	connections *Connections
}

func (a *Admin) Close() {
	a.sdk.Close()
}

func NewAdmin(config *Config) *Admin {
	connections, err := NewConnections(filepath.Clean(config.GatewayConfigPath))
	if err != nil {
		logger.Fatalf("failed to parse connections: %v", err)
	}

	sdk, err := fabsdk.New(cfg.FromFile(filepath.Clean(config.GatewayConfigPath)))
	if err != nil {
		logger.Fatalf("failed to create sdk: %v", err)
	}
	//defer sdk.Close()

	orgAdmin := "Admin"
	orgName := "org1"
	adminContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))

	client, err := fgosdkresmgmt.NewClient(adminContext)
	if err != nil {
		logger.Fatalf("failed to create context: %v", err)
	}

	return &Admin{sdk: sdk, client: client, config: config, connections: connections}
}

func (a *Admin) InitEnclave(targetPeer string) error {

	logger.Infof("--> Collection attestation params ")
	attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to load attestation params from environment: %v", err)
	}

	initReq := fpcmgmt.LifecycleInitEnclaveRequest{
		ChaincodeID:         a.config.ChaincodeId,
		EnclavePeerEndpoint: targetPeer, // define the peer where we wanna init our enclave
		AttestationParams:   attestationParams,
	}

	logger.Infof("--> LifecycleInitEnclave ")
	_, err = a.client.LifecycleInitEnclave(a.config.ChannelId, initReq)

	if err != nil {
		return err
	}

	return nil
}
