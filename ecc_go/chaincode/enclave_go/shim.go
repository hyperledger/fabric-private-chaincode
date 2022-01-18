/*
Copyright Riccardo Zappoli (riccardo.zappoli@unifr.ch)
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"crypto/sha256"
	"fmt"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type FpcStubInterface struct {
	stub  shim.ChaincodeStubInterface
	input *pb.ChaincodeInput
	rwset *readWriteSet
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

type readWriteSet struct {
	reads  map[string]read
	writes map[string]write
}

type read struct {
	kvread kvrwset.KVRead
	hash   []byte
}

type write struct {
	kvwrite kvrwset.KVWrite
}

func (rwset *readWriteSet) toFPCKVSet() *protos.FPCKVSet {
	fpcKVSet := &protos.FPCKVSet{
		RwSet: &kvrwset.KVRWSet{
			Reads:  []*kvrwset.KVRead{},
			Writes: []*kvrwset.KVWrite{},
		},
		ReadValueHashes: [][]byte{},
	}

	// fill with reads
	for _, read := range rwset.reads {
		fpcKVSet.RwSet.Reads = append(fpcKVSet.RwSet.Reads, &read.kvread)
		fpcKVSet.ReadValueHashes = append(fpcKVSet.ReadValueHashes, read.hash)
	}

	// fill with writes
	for _, write := range rwset.writes {
		fpcKVSet.RwSet.Writes = append(fpcKVSet.RwSet.Writes, &write.kvwrite)
	}

	return fpcKVSet
}

func newReadWriteSet() *readWriteSet {
	return &readWriteSet{
		reads:  make(map[string]read),
		writes: make(map[string]write),
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
	// TODO double check
	return f.stub.GetTxID()
}

func (f *FpcStubInterface) GetChannelID() string {
	// TODO double check
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
	return f.sep.DecryptState(encValue)
}

func (f *FpcStubInterface) GetPublicState(key string) ([]byte, error) {
	value, err := f.stub.GetState(key)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(value)

	f.rwset.reads[key] = read{
		kvread: kvrwset.KVRead{
			Key:     key,
			Version: nil,
		},
		hash: hash[:],
	}

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
	f.rwset.writes[key] = write{
		kvwrite: kvrwset.KVWrite{
			Key:      key,
			IsDelete: false,
			Value:    value,
		},
	}

	// note that since we are not using the fabric proposal response  we can skip the putState call
	//return f.stub.PutState(key, value)
	return nil
}

func (f *FpcStubInterface) DelState(key string) error {
	f.rwset.writes[key] = write{
		kvwrite: kvrwset.KVWrite{
			Key:      key,
			IsDelete: true,
		},
	}

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

// Quick and Dirty
// Really
type IteratorBuffer struct {
	buffer []*queryresult.KV
	index  int
}

func (i *IteratorBuffer) Add(in *queryresult.KV) {
	i.buffer = append(i.buffer, in)
}

func (i *IteratorBuffer) HasNext() bool {
	return i.index < len(i.buffer)
}

func (i *IteratorBuffer) Close() error {
	return nil
}

func (i *IteratorBuffer) Next() (*queryresult.KV, error) {
	out := i.buffer[i.index]
	i.index = i.index + 1
	return out, nil
}

func (f *FpcStubInterface) GetStateByPartialCompositeKey(objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	fmt.Println("Private start")
	fmt.Println("objectType", objectType)
	fmt.Println("keys", keys)
	iterator, err := f.GetPublicStateByPartialCompositeKey(objectType, keys)
	if err != nil {
		return nil, err
	}
	buffer := &IteratorBuffer{}
	for iterator.HasNext() {
		i, err := iterator.Next()
		if err != nil {
			return nil, err
		}
		decValue, err := f.sep.DecryptState(i.Value)
		if err != nil {
			return nil, err
		}
		b := &queryresult.KV{
			Namespace: i.Namespace,
			Key:       i.Key,
			Value:     decValue,
		}
		buffer.Add(b)
		fmt.Println(i.Key, "Key")
		fmt.Println(decValue, "DecValue")
	}
	fmt.Println("Private end")
	return buffer, nil
}

func (f *FpcStubInterface) GetPublicStateByPartialCompositeKey(objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	fmt.Println("Public start")
	fmt.Println("objectType", objectType)
	fmt.Println("keys", keys)
	iterator, err := f.stub.GetStateByPartialCompositeKey(objectType, keys)
	if err != nil {
		return nil, err
	}
	fmt.Println(&iterator)
	buffer := &IteratorBuffer{}
	for iterator.HasNext() {
		i, err := iterator.Next()
		if err != nil {
			return nil, err
		}
		v_hash := sha256.Sum256(i.Value)
		fmt.Println(i.Key, "Key")
		fmt.Println(utils.TransformToFPCKey(i.Key), "FPC")
		fmt.Println(i.Value, "Value")
		fmt.Println(v_hash)
		//f.fpcKvSet.RwSet.Reads = append(f.fpcKvSet.RwSet.Reads, &kvrwset.KVRead{
		//	Key:     utils.TransformToFPCKey(i.Key),
		//	Version: nil,
		//})
		//f.fpcKvSet.ReadValueHashes = append(f.fpcKvSet.ReadValueHashes, v_hash[:])
		b := &queryresult.KV{
			Namespace: i.Namespace,
			Key:       utils.TransformToFPCKey(i.Key),
			Value:     i.Value,
		}
		buffer.Add(b)
	}
	fmt.Println("Public end")
	return buffer, nil
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
	// TODO
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
	// TODO double check
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
	// TODO double check
	return f.stub.GetSignedProposal()
}

func (f *FpcStubInterface) GetTxTimestamp() (*timestamp.Timestamp, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FpcStubInterface) SetEvent(name string, payload []byte) error {
	panic("not implemented") // TODO: Implement
}
