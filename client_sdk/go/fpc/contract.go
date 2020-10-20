/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/pkg/errors"
)

type ContractInterface interface {
	ManagementInterface
	Name() string
	EvaluateTransaction(name string, args ...string) ([]byte, error)
	SubmitTransaction(name string, args ...string) ([]byte, error)
	//CreateTransaction(name string, opts ...gateway.TransactionOption) (*gateway.Transaction, error)
	RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error)
	Unregister(registration fab.Registration)
}

func GetContract(network *gateway.Network, chaincodeId string) ContractInterface {
	contract := network.GetContract(chaincodeId)
	ercc := network.GetContract("ercc")
	return &Contract{contract, ercc, nil, nil}
}

type Contract struct {
	contract                     *gateway.Contract
	ercc                         *gateway.Contract
	cachedChaincodeEncryptionKey []byte
	enclavePeers                 []string
}

func (c *Contract) Name() string {
	return c.contract.Name()
}

func (c *Contract) getChaincodeEncryptionKey() ([]byte, error) {
	if c.cachedChaincodeEncryptionKey == nil {
		ccKeyBytes, err := c.ercc.EvaluateTransaction("queryChaincodeEncryptionKey", c.Name())
		if err != nil {
			return nil, err
		}
		c.cachedChaincodeEncryptionKey = ccKeyBytes
	}
	return c.cachedChaincodeEncryptionKey, nil
}

func (c *Contract) getEnclavePeers() ([]string, error) {
	if c.enclavePeers == nil {
		// TODO: implement me to support multi-peer scenarios (currently createEnclave also populates c.enclavePeers ...
	}
	return c.enclavePeers, nil
}

// TODO better move to TX. crypto lib?! TBD
func (c *Contract) prepareChaincodeInvocation(name string, args []string, resultEncryptionKey []byte) (string, error) {
	p := &utils.ChaincodeParams{
		Function:            name,
		Args:                args,
		ResultEncryptionKey: resultEncryptionKey,
	}

	pBytes, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	log.Printf("prepping chaincode params: %s\n", p)

	k, err := c.getChaincodeEncryptionKey()
	if err != nil {
		return "", err
	}

	encryptedParams, err := Encrypt(pBytes, k)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedParams), nil
}

func (c *Contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {

	// pick response encryption key
	resultEncryptionKey, err := KeyGen()
	if err != nil {
		return nil, err
	}

	encryptedParamsBase64, err := c.prepareChaincodeInvocation(name, args, resultEncryptionKey)
	if err != nil {
		return nil, err
	}

	// note that WithEndorsingPeers is only used with txn.Submit!!!
	// GO SDK needs to be patched! We should create a PR for that!
	txn, err := c.contract.CreateTransaction(
		"__invoke",
		gateway.WithEndorsingPeers(c.enclavePeers...),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("calling __invoke!\n")
	responseBytes, err := txn.Evaluate(encryptedParamsBase64)
	if err != nil {
		return nil, err
	}

        // TODO maybe move this to a sub-function like prepareChaincodeInvocation?! TBD
	response, err := utils.UnmarshalResponse(responseBytes)
	if err != nil {
		return nil, err
	}

	// decrypt result
	return Decrypt(response.ResponseData, resultEncryptionKey)
}

func (c *Contract) SubmitTransaction(name string, args ...string) ([]byte, error) {

	// pick response encryption key
	resultEncryptionKey, err := KeyGen()
	if err != nil {
		return nil, err
	}

	encryptedParamsBase64, err := c.prepareChaincodeInvocation(name, args, resultEncryptionKey)
	if err != nil {
		return nil, err
	}

	txn, err := c.contract.CreateTransaction(
		"__invoke",
		gateway.WithEndorsingPeers(c.enclavePeers...),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("calling __invoke!\n")
	//responseBytes, err := c.contract.EvaluateTransaction("__invoke", encryptedParamsBase64)
	responseBytes, err := txn.Evaluate(encryptedParamsBase64)

	// first invoke (query) fpc chaincode
	//responseBytes, err := c.contract.EvaluateTransaction("__invoke", encryptedParamsBase64)
	if err != nil {
		return nil, errors.Wrap(err, "evaluation transaction failed")
	}

	log.Printf("calling __endorse!\n")
	// next invoke chaincode endorsement
	_, err = c.contract.SubmitTransaction("__endorse", base64.StdEncoding.EncodeToString(responseBytes))
	if err != nil {
		return nil, err
	}

	response, err := utils.UnmarshalResponse(responseBytes)
	if err != nil {
		return nil, err
	}

	return Decrypt(response.ResponseData, resultEncryptionKey)
}

//func (c *Contract) CreateTransaction(name string, opts ...gateway.TransactionOption) (*gateway.Transaction, error) {
//	return c.CreateTransaction(name, opts...)
//}

func (c *Contract) RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	return c.contract.RegisterEvent(eventFilter)
}

func (c *Contract) Unregister(registration fab.Registration) {
	c.contract.Unregister(registration)
}
