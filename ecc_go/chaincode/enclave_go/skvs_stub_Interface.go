/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"encoding/hex"
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

var SingleKey = "SingleKey"

type SkvsStubInterface struct {
	*FpcStubInterface
}

func NewSkvsStubInterface(stub shim.ChaincodeStubInterface, input *pb.ChaincodeInput, rwset *readWriteSet, sep StateEncryptionFunctions) *SkvsStubInterface {
	fpcStub := NewFpcStubInterface(stub, input, rwset, sep)
	return &SkvsStubInterface{fpcStub}
}

func (s *SkvsStubInterface) GetState(key string) ([]byte, error) {
	logger.Warningf("Calling Get State (Start), key = %s", key)
	encValue, err := s.GetPublicState(SingleKey)
	if err != nil {
		return nil, err
	}

	// in case the key does not exist, return early
	if len(encValue) == 0 {
		logger.Warningf("Calling Get State (End), data empty return.")
		return nil, nil
	}

	value, err := s.sep.DecryptState(encValue)
	if err != nil {
		return nil, err
	}

	// Create an interface{} value to hold the unmarshalled data
	var allData map[string]string

	// Unmarshal the JSON data into the interface{} value
	err = json.Unmarshal(value, &allData)
	if err != nil {
		logger.Errorf("SKVS Json unmarshal error: %s", err)
		return nil, err
	}
	logger.Warningf("Calling Get State (Mid), key = %s, decrypted done alldata = %s", key, allData)

	targetHex := allData[key]
	targetBytes, err := hex.DecodeString(targetHex)
	logger.Warningf("Calling Get State (End), Target: %s, TargetBytes: %x, err: %s", targetHex, targetBytes, err)
	return targetBytes, err
	// return s.sep.DecryptState(encValue)
}

func (s *SkvsStubInterface) PutState(key string, value []byte) error {
	logger.Warningf("Calling Put State (Start), key = %s, value = %x", key, value)
	// grab all data from the state.
	encAllValue, err := s.GetPublicState(SingleKey)
	if err != nil {
		return err
	}
	var allData map[string]string

	if len(encAllValue) == 0 {
		// current world state is empty
		allData = map[string]string{}
	} else {
		allValue, err := s.sep.DecryptState(encAllValue)
		if err != nil {
			return err
		}
		// Unmarshal the JSON data into the interface{} value
		err = json.Unmarshal(allValue, &allData)
		if err != nil {
			logger.Errorf("SKVS Json unmarshal error: %s", err)
			return err
		}
	}
	logger.Warningf("Calling Put State (Mid-1), decrypt succeed, allData = %s", allData)

	valueHex := hex.EncodeToString(value)

	allData[key] = valueHex
	logger.Warningf("Calling Put State (Mid-2), add need data key = %s, valueHex = %s, allData = %s", key, valueHex, allData)

	byteAllData, err := json.Marshal(allData)
	if err != nil {
		return err
	}
	logger.Warningf("Calling Put State (Mid-3), successfull marshal allData, byteAlldata = %x", byteAllData)

	encValue, err := s.sep.EncryptState(byteAllData)
	if err != nil {
		return err
	}
	logger.Warningf("Calling Put State (End), put encValue %x", encValue)

	return s.PutPublicState(SingleKey, encValue)
}
