/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

func NewSkvsStub(cc shim.Chaincode) *EnclaveStub {
	enclaveStub := NewEnclaveStub(cc)
	enclaveStub.stubProvider = func(stub shim.ChaincodeStubInterface, input *pb.ChaincodeInput, rwset *readWriteSet, sep StateEncryptionFunctions) shim.ChaincodeStubInterface {
		return NewSkvsStubInterface(stub, input, rwset, sep)
	}
	return enclaveStub
}
