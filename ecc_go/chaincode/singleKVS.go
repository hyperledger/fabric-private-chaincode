/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode"
)

var SingleKey = "SingleKey"

type SKVSWrapper struct {
	*chaincode.EnclaveChaincode
}

type skvsStub struct {
	shim.ChaincodeStubInterface
}

func (s *skvsStub) GetState(key string) ([]byte, error) {
	fmt.Printf("Inside SKVS solution, GetState, key=%s\n", key)
	return s.ChaincodeStubInterface.GetState(SingleKey)
	// return s.ChaincodeStubInterface.GetState(key)
}

func (s *skvsStub) PutState(key string, value []byte) error {
	fmt.Printf("Inside SKVS solution, PutState, key=%s, value=%x\n", key, value)
	return s.ChaincodeStubInterface.PutState(SingleKey, value)
	// return s.ChaincodeStubInterface.PutState(key, value)
}

func (s *SKVSWrapper) GetStub() shim.ChaincodeStubInterface {
	// get the original stub
	stub := s.GetStub()
	fmt.Println("Inside SKVS solution, GetStub")
	// create a new stub with the overridden GetState() function
	skvsStub := &skvsStub{stub}
	return skvsStub
}
