/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpcUtils

import (
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	fpcmgmt "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"google.golang.org/grpc"
)

type Admin struct {
	gw     *client.Gateway
	conn   *grpc.ClientConn
	client *fpcmgmt.Client
	config *Config
}

func (a *Admin) Close() {
	a.gw.Close()
	a.conn.Close()
}

func NewAdmin(config *Config) *Admin {
	gw, conn, err := connectGateway(config)
	if err != nil {
		logger.Fatalf("failed to connect to gateway: %v", err)
	}

	client, err := fpcmgmt.New(gw)
	if err != nil {
		logger.Fatalf("failed to create context: %v", err)
	}

	return &Admin{gw: gw, conn: conn, client: client, config: config}
}

func (a *Admin) InitEnclave(targetPeer string) error {
	logger.Infof("--> Collection attestation params ")
	attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to load attestation params from environment: %w", err)
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
