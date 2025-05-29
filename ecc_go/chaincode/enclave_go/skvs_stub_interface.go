/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

const SKVSKey = "SKVS"

type SkvsStubInterface struct {
	*FpcStubInterface
	allDataOld map[string][]byte
	allDataNew map[string][]byte
	key        string
}

func NewSkvsStubInterface(stub shim.ChaincodeStubInterface, input *pb.ChaincodeInput, rwset *readWriteSet, sep StateEncryptionFunctions) *SkvsStubInterface {
	fpcStub := NewFpcStubInterface(stub, input, rwset, sep)
	skvsStub := &SkvsStubInterface{
		FpcStubInterface: fpcStub,
		allDataOld:       make(map[string][]byte),
		allDataNew:       make(map[string][]byte),
		key:              SKVSKey,
	}
	err := skvsStub.initSKVS()
	if err != nil {
		panic(fmt.Sprintf("Initializing SKVS failed, err: %v", err))
	}
	return skvsStub
}

func (s *SkvsStubInterface) initSKVS() error {

	// get current state, this will only operate once
	encValue, err := s.GetPublicState(s.key)
	if err != nil {
		return err
	}

	// return if the key initially does not exist
	if len(encValue) == 0 {
		logger.Warningf("SKVS is empty, Initiating.")
		return nil
	}

	value, err := s.sep.DecryptState(encValue)
	if err != nil {
		return err
	}
	logger.Debug("SKVS has default value, loading current value.")

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
	return nil
}

func (s *SkvsStubInterface) GetState(key string) ([]byte, error) {
	value, found := s.allDataOld[key]
	if !found {
		logger.Errorf("skvs allDataOld key: %s, not found", key)
		return nil, nil
	}
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
