package fpc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/fpc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric/protoutil"
)

type ManagementInterface interface {
	CreateEnclave(peer string, attestationParams ...string) error
}

type ManagementAPI struct {
	network *gateway.Network
}

func (c *Contract) CreateEnclave(peer string, attestationParams ...string) error {

	p, err := json.Marshal(&utils.AttestationParams{Params: attestationParams})
	attestationParamsBase64 := base64.StdEncoding.EncodeToString(p)
	log.Printf("Prep attestation params: %s\n", attestationParamsBase64)

	txn, err := c.contract.CreateTransaction(
		"__initEnclave",
		gateway.WithEndorsingPeers(peer),
	)
	if err != nil {
		return err
	}

	credentialsBytes, err := txn.Evaluate(attestationParamsBase64)
	if err != nil {
		return fmt.Errorf("evaluation error: %s", err)
	}

	credentials := &protos.Credentials{}
	err = proto.Unmarshal(credentialsBytes, credentials)
	if err != nil {
		return fmt.Errorf("cannot unmarshal credentials: %s", err)
	}
	log.Printf("Received credentials from enclave: %s\n", credentials)

	// perform attestation evidence transformation
	credentials, err = attestation.ToEvidence(credentials)
	if err != nil {
		return err
	}

	credentialsBytes = protoutil.MarshalOrPanic(credentials)
	credentialsBase64 := base64.StdEncoding.EncodeToString(credentialsBytes)

	log.Printf("Call registerEnclave at ERCC: %s\n", attestationParamsBase64)
	_, err = c.ercc.SubmitTransaction("registerEnclave", credentialsBase64)
	if err != nil {
		return err
	}

	c.enclavePeers = append(c.enclavePeers, peer)

	return nil
}
