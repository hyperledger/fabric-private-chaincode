/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resmgmt provides FPC specific chaincode management functionality.
//
// For more information on the FPC management commands and related constraints on chaincode versions and endorsement policies,
// see `$FPC_PATH/docs/design/fabric-v2+/fpc-management.md`
//
// Example:
//
//  adminContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
//
//  client, err := resmgmt.New(adminContext)
//  if err != nil {
//  	log.Fatal(err)
//  }
//
//  attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
//  if err != nil {
//  	log.Fatal(err)
//  }
//
//  initReq := resmgmt.LifecycleInitEnclaveRequest{
//  	ChaincodeID:         "my-fpc-chaincode",
//  	EnclavePeerEndpoint: "mypeer.myorg.example.com", // define the peer where we wanna init our enclave
//  	AttestationParams:   attestationParams,
//  }
//
//  initTxId, err := client.LifecycleInitEnclave("mychannel", initReq)
//  if err != nil {
//  	log.Fatal(err)
//  }
//
// See also `lifecycle_test.go` and `$FPC_PATH/integration/client_sdk/go/utils/utils.go`
// for a running example.
//
package resmgmt

import (
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/lifecycle"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// LifecycleInitEnclaveRequest contains init enclave request parameters.
// In particular, it contains the FPC chaincode ID, the endpoint of the target peer to spawn the enclave, and
// attestation params to perform attestation and enclave registration.
type LifecycleInitEnclaveRequest struct {
	ChaincodeID         string
	EnclavePeerEndpoint string
	AttestationParams   *sgx.AttestationParams
}

// Client enables managing resources in Fabric network.
// It extends resmgmt.Client (https://pkg.go.dev/github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt#Client)
// from the standard Fabric Client SDK with additional FPC-specific functionality.
type Client struct {
	*resmgmt.Client
	lifecycleClient *lifecycle.Client
}

// New returns a FPC resource management client instance.
func New(ctxProvider context.ClientProvider, opts ...resmgmt.ClientOption) (*Client, error) {
	// get resource management client
	client, err := resmgmt.New(ctxProvider, opts...)
	if err != nil {
		return nil, err
	}

	lifecycleClient, err := lifecycle.New(NewChannelClientProvider(ctxProvider).ChannelClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:          client,
		lifecycleClient: lifecycleClient,
	}, nil
}

// LifecycleInitEnclave initializes and registers an enclave for a particular FPC chaincode.
func (rc *Client) LifecycleInitEnclave(channelId string, req LifecycleInitEnclaveRequest, options ...resmgmt.RequestOption) (fab.TransactionID, error) {
	txID, err := rc.lifecycleClient.LifecycleInitEnclave(channelId, lifecycle.LifecycleInitEnclaveRequest{
		ChaincodeID:         req.ChaincodeID,
		EnclavePeerEndpoint: req.EnclavePeerEndpoint,
		AttestationParams:   req.AttestationParams,
	})
	if err != nil {
		return fab.EmptyTransactionID, err
	}
	return fab.TransactionID(txID), nil
}
