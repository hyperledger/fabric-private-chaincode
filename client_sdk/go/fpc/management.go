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

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/fpc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	pbatt "github.com/hyperledger-labs/fabric-private-chaincode/internal/protos/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric/protoutil"
)

type ManagementInterface interface {
	InitEnclave(peerEndpoint string, attestationParams ...string) error
}

type ManagementAPI struct {
	contract *gateway.Contract
	ercc     *gateway.Contract
}

func GetManagementAPI(network *gateway.Network, chaincodeId string) ManagementInterface {
	contract := network.GetContract(chaincodeId)
	ercc := network.GetContract("ercc")
	return &ManagementAPI{contract: contract, ercc: ercc}
}

func (c *ManagementAPI) InitEnclave(peerEndpoint string, attestationParams ...string) error {
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
	serializedJsonParams, err := json.Marshal(&utils.AttestationParams{Params: attestationParams})
	if err != nil {
		shim.Error(err.Error())
	}

	initMsg := &protos.InitEnclaveMessage{
		PeerEndpoint:      peerEndpoint,
		AttestationParams: protoutil.MarshalOrPanic(&pbatt.AttestationParameters{Parameters: serializedJsonParams}),
	}

	log.Printf("calling __initEnclave\n")
	credentialsBytes, err := txn.Evaluate(utils.ProtoAsBase64(initMsg))
	if err != nil {
		return fmt.Errorf("evaluation error: %s", err)
	}

	credentials := &protos.Credentials{}
	err = proto.Unmarshal(credentialsBytes, credentials)
	if err != nil {
		return fmt.Errorf("cannot unmarshal credentials: %s", err)
	}

	// perform attestation evidence transformation
	credentials, err = attestation.ToEvidence(credentials)
	if err != nil {
		return err
	}

	log.Printf("calling registerEnclave\n")
	_, err = c.ercc.SubmitTransaction("registerEnclave", utils.ProtoAsBase64(credentials))
	if err != nil {
		return err
	}

	return nil
}
