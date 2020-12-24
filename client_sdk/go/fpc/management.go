/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/fpc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// ManagementAPI provides FPC specific chaincode management functionality.
// ManagementAPI objects should be created using the GetManagementAPI() factory method.
// For an example of its use, see https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/client_sdk/go/test/main.go
// For more information on the FPC management commands and related constraints on chaincode versions and endorsement policies,
// see https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/client_sdk/docs/design/fabric-v2+/fpc-management.md
type ManagementAPI interface {
	// InitEnclave initializes and registers an enclave for a particular chaincode.
	//  Parameters:
	//  peerEndpoint is the endpoint on which the enclave should be instantiated.
	//  attestationParams are parameters used during attestation of the instantiated enclave.
	// For SGX, it expects that the `SGX_MODE` environment variable is properly defined.
	// Additionally, if `SGX_MODE` is `HW`, then also the `SGX_CREDENTIALS_PATH` environment
	// variable must be defined and point to a directory containing the Intel IAS credential
	// files `api_key.txt`, `spid.txt` and `spid_type.txt`. (See `${FPC_PATH}/README.md` for
	// more information on these files)
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

	// Set attestation paramaters
	type Params struct {
		AttestationType string `json:"attestation_type"`
		HexSpid         string `json:"hex_spid"`
		SigRL           string `json:"sig_rl"`
	}
	var params Params

	switch sgxMode := os.Getenv("SGX_MODE"); sgxMode {
	case "HW":
		sgxCredentialsPath := os.Getenv("SGX_CREDENTIALS_PATH")
		if sgxCredentialsPath == "" {
			return fmt.Errorf("SGX_CREDENTIALS_PATH environment variable undefined")
		}
		hexSpidPath := filepath.Join(sgxCredentialsPath, "spid.txt")
		hexSpid, err := ioutil.ReadFile(hexSpidPath)
		if err != nil {
			return fmt.Errorf("Could not read properly (hex) spid file %s: %v", hexSpidPath, err)
		}
		spidTypePath := filepath.Join(sgxCredentialsPath, "spid_type.txt")
		spidType, err := ioutil.ReadFile(spidTypePath)
		if err != nil {
			return fmt.Errorf("Could not read properly (hex) spid file %s: %v", spidTypePath, err)
		}

		params = Params{AttestationType: strings.TrimSuffix(string(spidType), "\n"), HexSpid: strings.TrimSuffix(string(hexSpid), "\n"), SigRL: ""}
	case "SIM":
		params = Params{AttestationType: "simulated"}
	default:
		return fmt.Errorf("SGX_MODE environment variable ill-defined: '%s'", sgxMode)
	}
	serializedJSONParams, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("Cannot marshall (json) attestation params '%v': %v", params, err)
	}
	logger.Debugf("found attestation params: '%v' (json='%s')", params, serializedJSONParams)

	initMsg := &protos.InitEnclaveMessage{
		PeerEndpoint:      peerEndpoint,
		AttestationParams: []byte(base64.StdEncoding.EncodeToString(serializedJSONParams)),
	}

	logger.Debugf("calling __initEnclave (%v)", initMsg)
	credentialsBytes, err := txn.Evaluate(utils.MarshallProto(initMsg))
	if err != nil {
		return fmt.Errorf("evaluation error: %s", err)
	}

	var convertedCredentials string
	convertedCredentials, err = ConvertCredentials(string(credentialsBytes))
	if err != nil {
		return fmt.Errorf("evaluation error: %s", err)
	}

	logger.Debugf("calling registerEnclave")
	_, err = c.ercc.SubmitTransaction("registerEnclave", convertedCredentials)
	if err != nil {
		return err
	}

	return nil
}

// perform attestation evidence transformation
func ConvertCredentials(credentialsOnlyAttestation string) (credentialsWithEvidence string, err error) {
	logger.Debugf("Received Credential: '%s'", credentialsOnlyAttestation)
	credentials, err := utils.UnmarshalCredentials(credentialsOnlyAttestation)
	if err != nil {
		return "", fmt.Errorf("cannot decode credentials: %s", err)
	}

	credentials, err = attestation.ToEvidence(credentials)
	if err != nil {
		return "", err
	}
	credentialsOnlyAttestation = utils.MarshallProto(credentials)
	logger.Debugf("Converted to Credential: '%s'", credentialsOnlyAttestation)
	return credentialsOnlyAttestation, nil
}
