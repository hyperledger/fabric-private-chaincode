/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ercc

import (
	"errors"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// EnclaveRegistryStub interface
type EnclaveRegistryStub interface {
	GetSPID(stub shim.ChaincodeStubInterface, chaincodeName, channel string) ([]byte, error)
	RegisterEnclave(stub shim.ChaincodeStubInterface, chaincodeName, channel string, enclavePk, enclaveQuote []byte) error
}

// EnclaveRegistryStubImpl implements EnclaveRegistry interface and calls ercc
type EnclaveRegistryStubImpl struct {
}

// GetSPID return SPID from ercc
func (t *EnclaveRegistryStubImpl) GetSPID(stub shim.ChaincodeStubInterface, chaincodeName, channel string) ([]byte, error) {
	if spid, ok := stub.GetDecorations()["SPID"]; ok {
		return spid[:], nil
	}
	return nil, errors.New("Can not load SPID")
}

// RegisterEnclave registers enclave at ercc
func (t *EnclaveRegistryStubImpl) RegisterEnclave(stub shim.ChaincodeStubInterface, chaincodeName, channel string, enclavePk, enclaveQuote []byte) error {
	apiKey, ok := stub.GetDecorations()["apiKey"]
	if !ok {
		return errors.New("Can not load api-key")
	}

	resp := stub.InvokeChaincode(chaincodeName, [][]byte{[]byte("registerEnclave"), enclavePk, enclaveQuote, apiKey}, channel)
	if resp.Status != shim.OK {
		return errors.New("Setup failed: Can not register enclave at ercc: " + string(resp.Message))
	}
	return nil
}
