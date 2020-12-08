/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/fpc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	pbatt "github.com/hyperledger-labs/fabric-private-chaincode/internal/protos/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric/protoutil"
)

// ManagementAPI provides FPC specific chaincode management functionality.
// ManagementAPI objects should be created using the GetManagementAPI() factory method.
// For an example of its use, see https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/client_sdk/go/test/main.go
type ManagementAPI interface {
	// InitEnclave initializes and registers an enclave for a particular chaincode.
	//  Parameters:
	//  peerEndpoint is the endpoint on which the enclave should be instantiated.
	//  attestationParams are parameters used during attestation of the instantiated enclave.
	InitEnclave(peerEndpoint string, attestationParams ...string) error
}

// GetManagementAPI is the factory method for ManagementAPI objects.
//  Parameters:
//  network is an initialized Fabric network object
//  chaincodeID is the ID of the target chaincode
//
//  Returns:
//  The ManagementAPI object
func GetManagementAPI(network *gateway.Network, chaincodeID string) ManagementAPI {
	contract := network.GetContract(chaincodeID)
	ercc := network.GetContract("ercc")
	return &managementState{contract: contract, ercc: ercc}
}

type managementState struct {
	contract *gateway.Contract
	ercc     *gateway.Contract
}

func (c *managementState) InitEnclave(peerEndpoint string, attestationParams ...string) error {
	txn, err := c.contract.CreateTransaction(
		"__initEnclave",
		gateway.WithEndorsingPeers(peerEndpoint),
	)
	if err != nil {
		return err
	}

	if err := utils.ValidateEndpoint(peerEndpoint); err != nil {
		return err
	}

	// TODO revisit this once it becomes clear what attestationParams ...string is here
	serializedJSONParams, err := json.Marshal(&utils.AttestationParams{Params: attestationParams})
	if err != nil {
		shim.Error(err.Error())
	}

	initMsg := &protos.InitEnclaveMessage{
		PeerEndpoint:      peerEndpoint,
		AttestationParams: protoutil.MarshalOrPanic(&pbatt.AttestationParameters{Parameters: serializedJSONParams}),
	}

	log.Printf("calling __initEnclave\n")
	credentialsBytes, err := txn.Evaluate(utils.MarshallProto(initMsg))
	if err != nil {
		return fmt.Errorf("evaluation error: %s", err)
	}

	var convertedCredentials string
	convertedCredentials, err = ConvertCredentials(string(credentialsBytes))
	if err != nil {
		return fmt.Errorf("evaluation error: %s", err)
	}

	log.Printf("calling registerEnclave\n")
	_, err = c.ercc.SubmitTransaction("registerEnclave", convertedCredentials)
	if err != nil {
		return err
	}

	return nil
}

// perform attestation evidence transformation
func ConvertCredentials(credentialsOnlyAttestation string) (credentialsWithEvidence string, err error) {
	log.Printf("Received Credential: '%s'", credentialsOnlyAttestation)
	credentials, err := utils.UnmarshalCredentials(credentialsOnlyAttestation)
	if err != nil {
		return "", fmt.Errorf("cannot decode credentials: %s", err)
	}

	credentials, err = attestation.ToEvidence(credentials)
	if err != nil {
		return "", err
	}
	credentialsOnlyAttestation = utils.MarshallProto(credentials)
	log.Printf("Converted to Credential: '%s'", credentialsOnlyAttestation)
	return credentialsOnlyAttestation, nil
}
