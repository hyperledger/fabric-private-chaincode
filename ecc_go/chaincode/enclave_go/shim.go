/*
Copyright Riccardo Zappoli (riccardo.zappoli@unifr.ch)
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type FpcStubInterface struct {
	stub  shim.ChaincodeStubInterface
	input *pb.ChaincodeInput
	rwset ReadWriteSet
	sep   StateEncryptionFunctions
}

func NewFpcStubInterface(stub shim.ChaincodeStubInterface, input *pb.ChaincodeInput, rwset *readWriteSet, sep StateEncryptionFunctions) *FpcStubInterface {
	return &FpcStubInterface{
		stub:  stub,
		input: input,
		sep:   sep,
		rwset: rwset,
	}
}

func (f *FpcStubInterface) GetArgs() [][]byte {
	// note that we extract the invocation arguments from the contents of the FPC invocation and not the ChaincodeStubInterface
	return f.input.GetArgs()
}

func (f *FpcStubInterface) GetStringArgs() []string {
	args := f.GetArgs()
	strargs := make([]string, 0, len(args))
	for _, barg := range args {
		strargs = append(strargs, string(barg))
	}
	return strargs
}

func (f *FpcStubInterface) GetFunctionAndParameters() (function string, params []string) {
	args := f.GetStringArgs()
	function = ""
	params = []string{}
	if len(args) >= 1 {
		function = args[0]
		params = args[1:]
	}
	return
}

func (f *FpcStubInterface) GetArgsSlice() ([]byte, error) {
	args := f.GetArgs()
	var res []byte
	for _, barg := range args {
		res = append(res, barg...)
	}
	return res, nil
}

func (f *FpcStubInterface) GetTxID() string {
	return f.stub.GetTxID()
}

func (f *FpcStubInterface) GetChannelID() string {
	return f.stub.GetChannelID()
}

func (f *FpcStubInterface) InvokeChaincode(chaincodeName string, args [][]byte, channel string) pb.Response {
	panic("not supported")
}

func (f *FpcStubInterface) GetState(key string) ([]byte, error) {
	encValue, err := f.GetPublicState(key)
	if err != nil {
		return nil, err
	}

	// in case the key does not exist, return early
	if len(encValue) == 0 {
		return nil, nil
	}

	return f.sep.DecryptState(encValue)
}

func (f *FpcStubInterface) GetPublicState(key string) ([]byte, error) {
	value, err := f.stub.GetState(key)
	if err != nil {
		return nil, err
	}

	f.rwset.AddRead(key, hash(value))

	return value, nil
}

func (f *FpcStubInterface) PutState(key string, value []byte) error {
	encValue, err := f.sep.EncryptState(value)
	if err != nil {
		return err
	}
	return f.PutPublicState(key, encValue)
}

func (f *FpcStubInterface) PutPublicState(key string, value []byte) error {
	f.rwset.AddWrite(key, value)

	// note that since we are not using the fabric proposal response  we can skip the putState call
	//return f.stub.PutState(key, value)
	return nil
}

func (f *FpcStubInterface) DelState(key string) error {
	f.rwset.AddDelete(key)

	// note that since we are not using the fabric proposal response  we can skip the delState call
	//return f.stub.DelState(key)
	return nil
}

func (f *FpcStubInterface) SetStateValidationParameter(key string, ep []byte) error {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetStateValidationParameter(key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetStateByRange(startKey string, endKey string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetStateByRangeWithPagination(startKey string, endKey string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetStateByPartialCompositeKey(objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	iterator, err := f.stub.GetStateByPartialCompositeKey(objectType, keys)
	if err != nil {
		return nil, err
	}

	return newFpcIterator(iterator, f.rwset.AddRead, f.sep.DecryptState), nil
}

func (f *FpcStubInterface) GetPublicStateByPartialCompositeKey(objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	iterator, err := f.stub.GetStateByPartialCompositeKey(objectType, keys)
	if err != nil {
		return nil, err
	}

	// note that we do not pass the state decryption function here
	return newFpcIterator(iterator, f.rwset.AddRead, nil), nil
}

func (f *FpcStubInterface) GetStateByPartialCompositeKeyWithPagination(objectType string, keys []string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	key, err := f.stub.CreateCompositeKey(objectType, attributes)
	if err != nil {
		return "", err
	}
	return utils.TransformToFPCKey(key), nil
}

func (f *FpcStubInterface) SplitCompositeKey(compositeKey string) (string, []string, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetQueryResultWithPagination(query string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetHistoryForKey(key string) (shim.HistoryQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetPrivateData(collection string, key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetPrivateDataHash(collection string, key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) PutPrivateData(collection string, key string, value []byte) error {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) DelPrivateData(collection string, key string) error {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) SetPrivateDataValidationParameter(collection string, key string, ep []byte) error {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetPrivateDataValidationParameter(collection string, key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetPrivateDataByRange(collection string, startKey string, endKey string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetPrivateDataByPartialCompositeKey(collection string, objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetPrivateDataQueryResult(collection string, query string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetCreator() ([]byte, error) {
	return f.stub.GetCreator()
}

func (f *FpcStubInterface) GetTransient() (map[string][]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetBinding() ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetDecorations() map[string][]byte {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) GetSignedProposal() (*pb.SignedProposal, error) {
	return f.stub.GetSignedProposal()
}

func (f *FpcStubInterface) GetTxTimestamp() (*timestamp.Timestamp, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) SetEvent(name string, payload []byte) error {
	panic("not implemented") // TODO: Implement
}
