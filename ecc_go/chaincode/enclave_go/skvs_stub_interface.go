/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

type SkvsStubInterface struct {
	*FpcStubInterface
	allDataOld map[string][]byte
	allDataNew map[string][]byte
	key        string
}

func NewSkvsStubInterface(stub shim.ChaincodeStubInterface, input *pb.ChaincodeInput, rwset *readWriteSet, sep StateEncryptionFunctions) *SkvsStubInterface {
	logger.Warning("==== Get New Skvs Interface =====")
	fpcStub := NewFpcStubInterface(stub, input, rwset, sep)
	skvsStub := SkvsStubInterface{fpcStub, map[string][]byte{}, map[string][]byte{}, "SKVS"}
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
		if err != nil {
			logger.Errorf("SKVS Json unmarshal error: %s", err)
			return err
		}
		err = json.Unmarshal(value, &s.allDataNew)
		if err != nil {
			logger.Errorf("SKVS Json unmarshal error: %s", err)
			return err
		}
	}

	logger.Warningf("SKVS Init finish, allDataOld: %s, allDataNew: %s", s.allDataOld, s.allDataNew)
	return nil
}

func (s *SkvsStubInterface) GetState(key string) ([]byte, error) {
	logger.Warningf("Calling Get State (Start), key: %s, alldataOld: %s", key, s.allDataOld)
	value, found := s.allDataOld[key]
	if !found {
		return nil, errors.New("skvs allDataOld key not found")
	}
	logger.Warningf("Calling Get State (End), key: %s, value: %x", key, value)
	return value, nil
}

func (s *SkvsStubInterface) PutState(key string, value []byte) error {
	logger.Warningf("Calling Put State (Start), key: %s, value: %x, alldata: %s", key, value, s.allDataNew)

	s.allDataNew[key] = value
	byteAllData, err := json.Marshal(s.allDataNew)
	if err != nil {
		return err
	}
	encValue, err := s.sep.EncryptState(byteAllData)
	if err != nil {
		return err
	}
	logger.Warningf("Calling Put State (End), put encValue: %x", encValue)

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
