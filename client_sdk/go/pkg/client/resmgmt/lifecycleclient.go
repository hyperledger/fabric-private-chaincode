/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resmgmt provides FPC specific chaincode management functionality.
//
// For more information on the FPC management commands and related constraints on chaincode versions and endorsement policies,
// see https://github.com/hyperledger/fabric-private-chaincode/blob/main/docs/design/fabric-v2+/fpc-management.md
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
// See also https://github.com/hyperledger/fabric-private-chaincode/blob/main/integration/client_sdk/go/utils.go
// for a running example.
//
package resmgmt

import (
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
)

const (
	ercc               = "ercc"
	initEnclaveCMD     = "__initEnclave"
	registerEnclaveCMD = "registerEnclave"
)

var logger = flogging.MustGetLogger("fpc-client-resmgmt")

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
	getChannelClient getChannelClientFunction
	converter        credentialConverter
}

// helper interfaces for better testing
type channelClient interface {
	Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	Execute(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	InvokeHandler(handler invoke.Handler, request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	RegisterChaincodeEvent(chainCodeID string, eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error)
	UnregisterChaincodeEvent(registration fab.Registration)
}
type getChannelClientFunction func(channelId string) (channelClient, error)

type credentialConverter interface {
	ConvertCredentials(credentialsOnlyAttestation string) (credentialsWithEvidence string, err error)
}

// New returns a FPC resource management client instance.
func New(ctxProvider context.ClientProvider, opts ...resmgmt.ClientOption) (*Client, error) {
	// get resource management client
	client, err := resmgmt.New(ctxProvider, opts...)
	if err != nil {
		return nil, err
	}

	// create getChannelClient function
	getChannelClient := func(channelId string) (channelClient, error) {
		channelProvider := func() (context.Channel, error) {
			return contextImpl.NewChannel(ctxProvider, channelId)
		}

		return channel.New(channelProvider)
	}

	// use default credentials converter
	// TODO allow to override credential converter using opts
	converter := attestation.NewCredentialConverter()

	return &Client{client, getChannelClient, converter}, nil
}

// LifecycleInitEnclave initializes and registers an enclave for a particular FPC chaincode.
func (rc *Client) LifecycleInitEnclave(channelId string, req LifecycleInitEnclaveRequest, options ...resmgmt.RequestOption) (fab.TransactionID, error) {
	err := rc.verifyInitEnclaveRequest(req)
	if err != nil {
		return fab.EmptyTransactionID, err
	}

	channelClient, err := rc.getChannelClient(channelId)
	if err != nil {
		return fab.EmptyTransactionID, errors.Wrap(err, "Failed to create new channel client")
	}

	// serialize provided attestation params
	serializedJSONParams, err := req.AttestationParams.ToBase64EncodedJSON()
	if err != nil {
		return fab.EmptyTransactionID, errors.Wrap(err, "Failed to serialize attestation parameters")
	}
	logger.Debugf("using attestation params: '%v'", req.AttestationParams)

	initMsg := &protos.InitEnclaveMessage{
		PeerEndpoint:      req.EnclavePeerEndpoint,
		AttestationParams: serializedJSONParams,
	}

	initRequest := channel.Request{
		ChaincodeID: req.ChaincodeID,
		Fcn:         initEnclaveCMD,
		Args:        [][]byte{[]byte(utils.MarshallProto(initMsg))},
	}

	var initOpts []channel.RequestOption
	initOpts = append(initOpts, channel.WithRetry(retry.Opts{Attempts: 0}))
	initOpts = append(initOpts, channel.WithTargetEndpoints(req.EnclavePeerEndpoint))

	logger.Debugf("calling __initEnclave (%v)", initMsg)
	// send query to create (init) enclave at the target peer
	initResponse, err := channelClient.Query(initRequest, initOpts...)
	if err != nil {
		return fab.EmptyTransactionID, errors.Wrap(err, "Failed to query init enclave")
	}

	// convert credentials received from enclave
	convertedCredentials, err := rc.converter.ConvertCredentials(string(initResponse.Payload))
	if err != nil {
		return fab.EmptyTransactionID, errors.Wrap(err, "credentials conversion error")
	}

	registerRequest := channel.Request{
		ChaincodeID: ercc,
		Fcn:         registerEnclaveCMD,
		Args:        [][]byte{[]byte(convertedCredentials)},
	}

	var registerOpts []channel.RequestOption
	// TODO translate `resmgmt.RequestOption` to `channel.Option` options so we can pass it to execute
	//registerOpts = append(registerOpts, options...)

	logger.Debugf("calling registerEnclave")
	// invoke registerEnclave at enclave registry
	registerResponse, err := channelClient.Execute(registerRequest, registerOpts...)
	if err != nil {
		return fab.EmptyTransactionID, errors.Wrap(err, "Failed to execute register enclave")
	}

	return registerResponse.TransactionID, nil
}

func (rc *Client) verifyInitEnclaveRequest(req LifecycleInitEnclaveRequest) error {
	if req.ChaincodeID == "" {
		return errors.New("chaincodeId is required")
	}

	if req.EnclavePeerEndpoint == "" {
		return errors.New("target peer, which spawns the enclave, is required")
	}

	if req.AttestationParams == nil {
		return errors.New("attestation params are required")
	}

	err := req.AttestationParams.Validate()
	if err != nil {
		return errors.Wrap(err, "attestation params are invalid")
	}

	return nil
}
