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
//	client, err := resmgmt.New(gw, adminSign, func(endpoint string) (grpc.ClientConnInterface, error) {
//		return gateway.NewGrpcConnection(endpoint, tlsRootCertPEM, "")
//	})
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
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/lifecycle"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
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
// Enclaves are initialized by calling the target peer directly; the enclave registration is
// submitted via the Fabric Gateway.
type Client struct {
	lifecycleClient *lifecycle.Client
}

// New returns a FPC resource management client instance.
//
//	Parameters:
//	gw is a Fabric Gateway, e.g., obtained via client.Connect
//	sign is the signing function of the gateway identity
//	dial returns a connection to the peer given by LifecycleInitEnclaveRequest.EnclavePeerEndpoint
func New(gw Gateway, sign identity.Sign, dial gateway.PeerDialer) (*Client, error) {
	if gw == nil || sign == nil || dial == nil {
		return nil, errors.New("invalid arguments, gateway, sign and dial are required")
	}

	lifecycleClient, err := lifecycle.New(NewChannelClientProvider(gw, sign, dial).ChannelClient)
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
