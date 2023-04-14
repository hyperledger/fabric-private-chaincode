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
	"github.com/pkg/errors"
)

type SkvsStubInterface struct {
	*FpcStubInterface
	allDataOld map[string]string
	allDataNew map[string]string
	key        string
}

func NewSkvsStubInterface(stub shim.ChaincodeStubInterface, input *pb.ChaincodeInput, rwset *readWriteSet, sep StateEncryptionFunctions) *SkvsStubInterface {
	logger.Warning("==== Get New Skvs Interface =====")
	fpcStub := NewFpcStubInterface(stub, input, rwset, sep)
	skvsStub := SkvsStubInterface{fpcStub, map[string]string{}, map[string]string{}, "SKVS"}
	err := skvsStub.InitSKVS()
	if err != nil {
		logger.Warningf("Error!! Initializing SKVS failed")
	}
	return &skvsStub
}

func (s *SkvsStubInterface) InitSKVS() error {
	logger.Warningf(" === Initializing SKVS === ")

	// get current state, this will only operate once
	encValue, err := s.GetPublicState(s.key)
	if err != nil {
		return nil
	}

	if len(encValue) == 0 {
		logger.Warningf("SKVS is empty, Initiating.")
	} else {
		value, err := s.sep.DecryptState(encValue)
		if err != nil {
			return err
		}
		logger.Warningf("SKVS has default value, loading current value.")

		err = json.Unmarshal(value, &s.allDataOld)
		err = json.Unmarshal(value, &s.allDataNew)
		if err != nil {
			logger.Errorf("SKVS Json unmarshal error: %s", err)
			return err
		}
	}

	logger.Warningf("SKVS Init finish, allDataOld = %s, allDataNew = %s", s.allDataOld, s.allDataNew)
	return nil
}

func (s *SkvsStubInterface) GetState(key string) ([]byte, error) {
	logger.Warningf("Calling Get State (Start), key = %s, alldataOld = %s", key, s.allDataOld)
	targetHex, found := s.allDataOld[key]
	if found != true {
		return nil, errors.New("skvs allDataOld key not found")
	}
	targetBytes, err := hex.DecodeString(targetHex)
	logger.Warningf("Calling Get State (End), TargetHex: %s, TargetBytes: %x, err: %s", targetHex, targetBytes, err)
	return targetBytes, err
}

func (s *SkvsStubInterface) PutState(key string, value []byte) error {
	logger.Warningf("Calling Put State (Start), key = %s, value = %x, alldata = %s", key, value, s.allDataNew)

	valueHex := hex.EncodeToString(value)
	s.allDataNew[key] = valueHex
	logger.Warningf("Calling Put State (Mid-1), add need data key = %s, valueHex = %s, allData = %s", key, valueHex, s.allDataNew)

	byteAllData, err := json.Marshal(s.allDataNew)
	if err != nil {
		return err
	}
	logger.Warningf("Calling Put State (Mid-2), successfull marshal allData, byteAlldata = %x", byteAllData)

	encValue, err := s.sep.EncryptState(byteAllData)
	if err != nil {
		return err
	}
	logger.Warningf("Calling Put State (End), put encValue %x", encValue)

	return s.PutPublicState(s.key, encValue)
}
