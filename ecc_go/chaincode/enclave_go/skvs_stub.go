/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type SkvsStubInterface struct {
	*FpcStubInterface
	allDataOld map[string][]byte
	allDataNew map[string][]byte
	key        string
}

func NewSkvsStubInterface(stub shim.ChaincodeStubInterface, input *pb.ChaincodeInput, rwset *readWriteSet, sep StateEncryptionFunctions) shim.ChaincodeStubInterface {
	logger.Debugf("===== Creating New Skvs Interface =====")
	fpcStub := NewFpcStubInterface(stub, input, rwset, sep)
	skvsStub := SkvsStubInterface{fpcStub.(*FpcStubInterface), map[string][]byte{}, map[string][]byte{}, "SKVS"}
	err := skvsStub.InitSKVS()
	if err != nil {
		logger.Warningf("Error!! Initializing SKVS failed")
	}
	return &skvsStub
}

func (s *SkvsStubInterface) InitSKVS() error {
	// get current state, this will only operate once
	encValue, err := s.GetPublicState(s.key)
	if err != nil {
		return nil
	}

	if len(encValue) == 0 {
		logger.Debugf("SKVS is empty, Initiating.")
	} else {
		value, err := s.sep.DecryptState(encValue)
		if err != nil {
			return err
		}
		logger.Debugf("SKVS has default value, loading current value.")

		err = json.Unmarshal(value, &s.allDataOld)
		err = json.Unmarshal(value, &s.allDataNew)
		if err != nil {
			logger.Errorf("SKVS Json unmarshal error: %s", err)
			return err
		}
	}
	return nil
}

func (s *SkvsStubInterface) GetState(key string) ([]byte, error) {
	value := s.allDataOld[key]
	return value, nil
}

func (s *SkvsStubInterface) PutState(key string, value []byte) error {
	s.allDataNew[key] = value
	byteAllData, err := json.Marshal(s.allDataNew)
	if err != nil {
		return err
	}
	encValue, err := s.sep.EncryptState(byteAllData)
	if err != nil {
		return err
	}

	return s.PutPublicState(s.key, encValue)
}

func (s *SkvsStubInterface) DelState(key string) error {
	delete(s.allDataNew, key)
	byteAllData, err := json.Marshal(s.allDataNew)
	if err != nil {
		return err
	}
	encValue, err := s.sep.EncryptState(byteAllData)
	if err != nil {
		return err
	}
	return s.PutPublicState(s.key, encValue)
}

func (s *SkvsStubInterface) GetStateByRange(startKey string, endKey string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

func (s *SkvsStubInterface) GetStateByRangeWithPagination(startKey string, endKey string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	panic("not implemented") // TODO: Implement
}
