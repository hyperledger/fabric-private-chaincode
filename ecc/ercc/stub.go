/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package ercc

import (
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
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
