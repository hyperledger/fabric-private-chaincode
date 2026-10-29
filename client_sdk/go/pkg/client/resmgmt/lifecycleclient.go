/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resmgmt provides FPC specific chaincode management functionality.
//
// For more information on the FPC management commands and related constraints on chaincode versions and endorsement policies,
// see `$FPC_PATH/docs/design/fabric-v2+/fpc-management.md`
//
// Chaincode lifecycle operations that are not FPC specific (install, approve, commit) are provided by
// the Fabric Admin SDK (https://pkg.go.dev/github.com/hyperledger/fabric-admin-sdk).
//
// Example:
//
//	gw, err := client.Connect(adminIdentity, client.WithSign(adminSign), client.WithClientConnection(conn))
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer gw.Close()
//
//	client, err := resmgmt.New(gw)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	initReq := resmgmt.LifecycleInitEnclaveRequest{
//		ChaincodeID:         "my-fpc-chaincode",
//		EnclavePeerEndpoint: "mypeer.myorg.example.com", // define the peer where we wanna init our enclave
//		AttestationParams:   attestationParams,
//	}
//
//	initTxId, err := client.LifecycleInitEnclave("mychannel", initReq)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// See also `lifecycle_test.go` and `$FPC_PATH/integration/client_sdk/go/utils/utils.go`
// for a running example.
package resmgmt

import (
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/lifecycle"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/pkg/errors"
)

// LifecycleInitEnclaveRequest contains init enclave request parameters.
// In particular, it contains the FPC chaincode ID, the endpoint of the target peer to spawn the enclave, and
// attestation params to perform attestation and enclave registration.
type LifecycleInitEnclaveRequest struct {
	ChaincodeID         string
	EnclavePeerEndpoint string
	AttestationParams   *sgx.AttestationParams
}

// Client enables FPC specific resource management in a Fabric network.
// It must operate on a connection to the Fabric Gateway of the target enclave peer,
// as the Fabric Gateway routes requests by organization rather than by peer.
type Client struct {
	lifecycleClient *lifecycle.Client
}

// New returns a FPC resource management client instance.
func New(gateway Gateway) (*Client, error) {
	if gateway == nil {
		return nil, errors.New("invalid arguments, gateway is nil")
	}

	lifecycleClient, err := lifecycle.New(NewChannelClientProvider(gateway).ChannelClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		lifecycleClient: lifecycleClient,
	}, nil
}

// LifecycleInitEnclave initializes and registers an enclave for a particular FPC chaincode.
// It returns the ID of the enclave registration transaction.
func (rc *Client) LifecycleInitEnclave(channelId string, req LifecycleInitEnclaveRequest) (string, error) {
	return rc.lifecycleClient.LifecycleInitEnclave(channelId, lifecycle.LifecycleInitEnclaveRequest{
		ChaincodeID:         req.ChaincodeID,
		EnclavePeerEndpoint: req.EnclavePeerEndpoint,
		AttestationParams:   req.AttestationParams,
	})
}
