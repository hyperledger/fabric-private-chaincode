/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package lifecycle provides FPC specific chaincode management functionality.
//
// For more information on the FPC management commands and related constraints on chaincode versions and endorsement policies,
// see https://github.com/hyperledger/fabric-private-chaincode/blob/main/docs/design/fabric-v2+/fpc-management.md
//
// Example:
//
//  adminContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
//
//  client, err := lifecycle.New(adminContext)
//  if err != nil {
//  	log.Fatal(err)
//  }
//
//  attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
//  if err != nil {
//  	log.Fatal(err)
//  }
//
//  initReq := lifecycle.LifecycleInitEnclaveRequest{
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
package lifecycle

import (
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
)

const (
	ERCC               = "ercc"
	InitEnclaveCMD     = "__initEnclave"
	RegisterEnclaveCMD = "registerEnclave"
)

var logger = flogging.MustGetLogger("fpc-client-lifecycle")

// LifecycleInitEnclaveRequest contains init enclave request parameters.
// In particular, it contains the FPC chaincode ID, the endpoint of the target peer to spawn the enclave, and
// attestation params to perform attestation and enclave registration.
type LifecycleInitEnclaveRequest struct {
	ChaincodeID         string
	EnclavePeerEndpoint string
	AttestationParams   *sgx.AttestationParams
}

type CredentialConverter interface {
	ConvertCredentials(credentialsOnlyAttestation string) (credentialsWithEvidence string, err error)
}

// ChannelClient models an interface to query and execute chaincodes
type ChannelClient interface {
	Query(chaincodeID string, fcn string, args [][]byte, targetEndpoints ...string) ([]byte, error)
	Execute(chaincodeID string, fcn string, args [][]byte) (string, error)
}

type GetChannelClientFunction func(channelID string) (ChannelClient, error)

// Client enables managing resources in Fabric network.
// It extends lifecycle.Client (https://pkg.go.dev/github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt#Client)
// from the standard Fabric Client SDK with additional FPC-specific functionality.
type Client struct {
	GetChannelClient GetChannelClientFunction
	Converter        CredentialConverter
}

// New returns a FPC resource management client instance.
func New(getChannelClient GetChannelClientFunction) (*Client, error) {
	// use default credentials converter
	// TODO allow to override credential converter using opts
	if getChannelClient == nil {
		return nil, errors.Errorf("invalid arguments, channel client loader is nil")
	}

	converter := attestation.NewDefaultCredentialConverter()
	return &Client{GetChannelClient: getChannelClient, Converter: converter}, nil
}

// LifecycleInitEnclave initializes and registers an enclave for a particular FPC chaincode.
func (rc *Client) LifecycleInitEnclave(channelID string, req LifecycleInitEnclaveRequest) (string, error) {
	err := rc.verifyInitEnclaveRequest(req)
	if err != nil {
		return "", err
	}

	channelClient, err := rc.GetChannelClient(channelID)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create new channel client")
	}

	// serialize provided attestation params
	serializedJSONParams, err := req.AttestationParams.ToBase64EncodedJSON()
	if err != nil {
		return "", errors.Wrap(err, "Failed to serialize attestation parameters")
	}
	logger.Debugf("using attestation params: '%v'", req.AttestationParams)

	initMsg := &protos.InitEnclaveMessage{
		PeerEndpoint:      req.EnclavePeerEndpoint,
		AttestationParams: serializedJSONParams,
	}

	// var initOpts []channel.RequestOption
	// initOpts = append(initOpts, channel.WithRetry(retry.Opts{Attempts: 0}))
	// initOpts = append(initOpts, channel.WithTargetEndpoints(req.EnclavePeerEndpoint))

	logger.Debugf("calling __initEnclave (%v)", initMsg)
	// send query to create (init) enclave at the target peer
	payload, err := channelClient.Query(
		req.ChaincodeID, InitEnclaveCMD, [][]byte{[]byte(utils.MarshallProtoBase64(initMsg))},
		req.EnclavePeerEndpoint,
	)
	if err != nil {
		return "", errors.Wrap(err, "Failed to query init enclave")
	}

	// convert credentials received from enclave
	convertedCredentials, err := rc.Converter.ConvertCredentials(string(payload))
	if err != nil {
		return "", errors.Wrap(err, "credentials conversion error")
	}

	logger.Debugf("calling registerEnclave")
	// invoke registerEnclave at enclave registry
	txID, err := channelClient.Execute(ERCC, RegisterEnclaveCMD, [][]byte{[]byte(convertedCredentials)})
	if err != nil {
		return "", errors.Wrap(err, "Failed to execute register enclave")
	}

	return txID, nil
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
