/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hyperledger-labs/fabric-private-chaincode/demo/client/backend/mock/api"

	"github.com/hyperledger-labs/fabric-private-chaincode/ecc"

	"github.com/hyperledger-labs/fabric-private-chaincode/utils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type Transaction struct {
	Uuid      string   `json:"uuid"`
	Chaincode string   `json:"chaincode"`
	Creator   string   `json:"creator"`
	Function  string   `json:"func"`
	Args      []string `json:"args"`
	Valid     bool     `json:"is_valid"`
	State     []byte   `json:"state"`
}

type MockStubWrapper struct {
	sync.RWMutex
	MockStub     *shim.MockStub
	Creator      string
	Seq          int
	Transactions []*Transaction
	notifier     *Notifier
	cc           shim.Chaincode
}

func NewWrapper(name string, cc shim.Chaincode, notifier *Notifier) *MockStubWrapper {
	stub := shim.NewMockStub(name, cc)
	return &MockStubWrapper{MockStub: stub, Seq: 0, notifier: notifier, cc: cc}
}

func DestroyChaincode(m *MockStubWrapper) {
	if e, ok := m.cc.(*ecc.EnclaveChaincode); ok {
		e.Destroy()
	}
}

func (m *MockStubWrapper) createUuid(uuid string) string {
	m.Seq++
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d_%s", m.Seq, uuid)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func isValidTransactions(resp peer.Response) bool {
	if resp.Status != shim.OK {
		return false
	}

	if resp.Payload != nil {
		var response utils.Response
		if err := json.Unmarshal(resp.Payload, &response); err != nil {
			return false
		}
		var responseObj ResponseObject
		if err := json.Unmarshal(response.ResponseData, &responseObj); err != nil {
			// what to do here?
		}

		if responseObj.Status.RC != 0 {
			return false
		}
	}

	return true
}

func (m *MockStubWrapper) saveTransaction(uuid string, resp peer.Response) {

	function, args := m.MockStub.GetFunctionAndParameters()

	// serialize our current State
	//s, _ := json.Marshal(m.MockStub.State)

	isValid := isValidTransactions(resp)
	// quick hack
	if function == "__setup" {
		function = "setup_enclave"
		isValid = true
	}

	if function == "__init" {
		function = "init_chaincode"
	}

	// new transaction
	t := &Transaction{
		Uuid:      uuid,
		Chaincode: m.MockStub.Name,
		Creator:   m.Creator,
		Args:      args,
		Function:  function,
		Valid:     isValid,
		//State:     s,
	}

	m.Transactions = append(m.Transactions, t)
	m.notifier.Submit(t)
}

// Initialise this chaincode,  also starts and ends a transaction.
func (m *MockStubWrapper) MockInit(uuid string, args [][]byte) pb.Response {
	var ok bool
	var mappedName api.MappedName

	if mappedName, ok = api.MockNameMap[m.Creator]; ok {
		logger.Debugf("Mapping user '%s' to { MspId: '%s', Org: '%s', User: '%s'}",
			m.Creator, mappedName.MspId, mappedName.Org, mappedName.User)
	} else {
		mappedName = api.MappedName{User: m.Creator, MspId: defaultMspId, Org: defaultOrg}
		logger.Debugf("No name mapping found for user '%s', using default MspId '%s' and Org '%s'",
			m.Creator, defaultMspId, defaultOrg)
	}

	creator, _ := generateMockCreator(mappedName.MspId, mappedName.Org, mappedName.User)
	m.MockStub.Creator = creator

	uuid = m.createUuid(uuid)
	resp := m.MockStub.MockInit(uuid, args)
	m.saveTransaction(uuid, resp)
	return resp
}

// Invoke this chaincode, also starts and ends a transaction.
func (m *MockStubWrapper) MockInvoke(uuid string, args [][]byte) pb.Response {
	var ok bool
	var mappedName api.MappedName

	if mappedName, ok = api.MockNameMap[m.Creator]; ok {
		logger.Debugf("Mapping user '%s' to { MspId: '%s', Org: '%s', User: '%s'}",
			m.Creator, mappedName.MspId, mappedName.Org, mappedName.User)
	} else {
		mappedName = api.MappedName{User: m.Creator, MspId: defaultMspId, Org: defaultOrg}
		logger.Debugf("No name mapping found for user '%s', using default MspId '%s' and Org '%s'",
			m.Creator, defaultMspId, defaultOrg)
	}

	creator, _ := generateMockCreator(mappedName.MspId, mappedName.Org, mappedName.User)
	m.MockStub.Creator = creator

	uuid = m.createUuid(uuid)
	resp := m.MockStub.MockInvoke(uuid, args)
	m.saveTransaction(uuid, resp)
	return resp
}

// Invoke this chaincode, also starts and ends a transaction.
func (m *MockStubWrapper) MockQuery(uuid string, args [][]byte) pb.Response {
	var ok bool
	var mappedName api.MappedName

	if mappedName, ok = api.MockNameMap[m.Creator]; ok {
		logger.Debugf("Mapping user '{}' to { MspId: '{}', Org: '{}', User: '{}'}",
			m.Creator, mappedName.MspId, mappedName.Org, mappedName.User)
	} else {
		mappedName = api.MappedName{User: m.Creator, MspId: defaultMspId, Org: defaultOrg}
		logger.Debugf("No name mapping found for user '{}', using default MspId '{}' and Org '{}'",
			m.Creator, defaultMspId, defaultOrg)
	}

	creator, _ := generateMockCreator(mappedName.MspId, mappedName.Org, mappedName.User)
	m.MockStub.Creator = creator

	// save state
	s, err := json.Marshal(m.MockStub.State)
	if err != nil {
		panic("error storing state")
	}

	uuid = m.createUuid(uuid)
	resp := m.MockStub.MockInvoke(uuid, args)

	// restore state
	var restoreState map[string][]byte
	err = json.Unmarshal(s, &restoreState)
	if err != nil {
		panic("error restoring state")
	}
	m.MockStub.State = restoreState

	return resp
}

func (m *MockStubWrapper) GetArgs() [][]byte {
	return m.MockStub.GetArgs()
}

func (m *MockStubWrapper) GetStringArgs() []string {
	return m.MockStub.GetStringArgs()
}

func (m *MockStubWrapper) GetFunctionAndParameters() (string, []string) {
	return m.MockStub.GetFunctionAndParameters()
}

func (m *MockStubWrapper) GetArgsSlice() ([]byte, error) {
	return m.MockStub.GetArgsSlice()
}

func (m *MockStubWrapper) GetTxID() string {
	return m.MockStub.GetTxID()
}

func (m *MockStubWrapper) GetChannelID() string {
	return m.MockStub.GetChannelID()
}

func (m *MockStubWrapper) InvokeChaincode(chaincodeName string, args [][]byte, channel string) peer.Response {
	return m.MockStub.InvokeChaincode(chaincodeName, args, channel)
}

func (m *MockStubWrapper) GetState(key string) ([]byte, error) {
	return m.MockStub.GetState(key)
}

func (m *MockStubWrapper) PutState(key string, value []byte) error {
	return m.MockStub.PutState(key, value)
}

func (m *MockStubWrapper) DelState(key string) error {
	return m.MockStub.DelState(key)
}

func (m *MockStubWrapper) SetStateValidationParameter(key string, ep []byte) error {
	panic("implement me")
}

func (m *MockStubWrapper) GetStateValidationParameter(key string) ([]byte, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetStateByRange(startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetStateByRangeWithPagination(startKey, endKey string, pageSize int32,
	bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetStateByPartialCompositeKey(objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	return m.MockStub.GetStateByPartialCompositeKey(objectType, keys)
}

func (m *MockStubWrapper) GetStateByPartialCompositeKeyWithPagination(objectType string, keys []string,
	pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	panic("implement me")
}

func (m *MockStubWrapper) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	return m.MockStub.CreateCompositeKey(objectType, attributes)
}

func (m *MockStubWrapper) SplitCompositeKey(compositeKey string) (string, []string, error) {
	return m.MockStub.SplitCompositeKey(compositeKey)
}

func (m *MockStubWrapper) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetQueryResultWithPagination(query string, pageSize int32,
	bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetHistoryForKey(key string) (shim.HistoryQueryIteratorInterface, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetPrivateData(collection, key string) ([]byte, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetPrivateDataHash(collection, key string) ([]byte, error) {
	panic("implement me")
}

func (m *MockStubWrapper) PutPrivateData(collection string, key string, value []byte) error {
	panic("implement me")
}

func (m *MockStubWrapper) DelPrivateData(collection, key string) error {
	panic("implement me")
}

func (m *MockStubWrapper) SetPrivateDataValidationParameter(collection, key string, ep []byte) error {
	panic("implement me")
}

func (m *MockStubWrapper) GetPrivateDataValidationParameter(collection, key string) ([]byte, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetPrivateDataByRange(collection, startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetPrivateDataByPartialCompositeKey(collection, objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetPrivateDataQueryResult(collection, query string) (shim.StateQueryIteratorInterface, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetCreator() ([]byte, error) {
	return m.MockStub.GetCreator()
}

func (m *MockStubWrapper) GetTransient() (map[string][]byte, error) {
	return m.MockStub.GetTransient()
}

func (m *MockStubWrapper) GetBinding() ([]byte, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetDecorations() map[string][]byte {
	panic("implement me")
}

func (m *MockStubWrapper) GetSignedProposal() (*peer.SignedProposal, error) {
	panic("implement me")
}

func (m *MockStubWrapper) GetTxTimestamp() (*interface{}, error) {
	panic("implement me")
}

func (m *MockStubWrapper) SetEvent(name string, payload []byte) error {
	panic("implement me")
}
