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
	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type FpcStubInterface struct {
	csp      crypto.CSP
	stub     shim.ChaincodeStubInterface
	input    *pb.ChaincodeInput
	fpcKvSet *protos.FPCKVSet
	stateKey []byte
}

func NewFpcStubInterface(stub shim.ChaincodeStubInterface, input *pb.ChaincodeInput, fpcKvSet *protos.FPCKVSet, stateKey []byte) *FpcStubInterface {
	return &FpcStubInterface{
		csp:      crypto.GetDefaultCSP(),
		stub:     stub,
		input:    input,
		fpcKvSet: fpcKvSet,
		stateKey: stateKey,
	}
}

// GetArgs returns the arguments intended for the chaincode Init and Invoke
// as an array of byte arrays.
func (f *FpcStubInterface) GetArgs() [][]byte {
	return f.input.GetArgs()
}

// GetStringArgs returns the arguments intended for the chaincode Init and
// Invoke as a string array. Only use GetStringArgs if the client passes
// arguments intended to be used as strings.
func (f *FpcStubInterface) GetStringArgs() []string {
	byteArgs := f.input.GetArgs()
	stringArgs := make([]string, len(byteArgs))
	for i := range byteArgs {
		stringArgs[i] = string(byteArgs[i])
	}
	return stringArgs
}

// GetFunctionAndParameters returns the first argument as the function
// name and the rest of the arguments as parameters in a string array.
// Only use GetFunctionAndParameters if the client passes arguments intended
// to be used as strings.
func (f *FpcStubInterface) GetFunctionAndParameters() (string, []string) {
	inputs := f.GetStringArgs()
	return inputs[0], inputs[1:]
}

// GetArgsSlice returns the arguments intended for the chaincode Init and
// Invoke as a byte array
func (f *FpcStubInterface) GetArgsSlice() ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// GetTxID returns the tx_id of the transaction proposal, which is unique per
// transaction and per client. See
// https://godoc.org/github.com/hyperledger/fabric-protos-go/common#ChannelHeader
// for further details.
func (f *FpcStubInterface) GetTxID() string {
	panic("not implemented") // TODO: Implement
}

// GetChannelID returns the channel the proposal is sent to for chaincode to process.
// This would be the channel_id of the transaction proposal (see
// https://godoc.org/github.com/hyperledger/fabric-protos-go/common#ChannelHeader )
// except where the chaincode is calling another on a different channel.
func (f *FpcStubInterface) GetChannelID() string {
	panic("not implemented") // TODO: Implement
}

// InvokeChaincode locally calls the specified chaincode `Invoke` using the
// same transaction context; that is, chaincode calling chaincode doesn't
// create a new transaction message.
// If the called chaincode is on the same channel, it simply adds the called
// chaincode read set and write set to the calling transaction.
// If the called chaincode is on a different channel,
// only the Response is returned to the calling chaincode; any PutState calls
// from the called chaincode will not have any effect on the ledger; that is,
// the called chaincode on a different channel will not have its read set
// and write set applied to the transaction. Only the calling chaincode's
// read set and write set will be applied to the transaction. Effectively
// the called chaincode on a different channel is a `Query`, which does not
// participate in state validation checks in subsequent commit phase.
// If `channel` is empty, the caller's channel is assumed.
func (f *FpcStubInterface) InvokeChaincode(chaincodeName string, args [][]byte, channel string) pb.Response {
	panic("not implemented") // TODO: Implement
}

// GetState returns the value of the specified `key` from the
// ledger. Note that GetState doesn't read data from the writeset, which
// has not been committed to the ledger. In other words, GetState doesn't
// consider data modified by PutState that has not been committed.
// If the key does not exist in the state database, (nil, nil) is returned.
func (f *FpcStubInterface) GetState(key string) ([]byte, error) {
	encValue, err := f.GetPublicState(key)
	if err != nil {
		return nil, err
	}
	value, err := f.csp.DecryptMessage(f.stateKey, encValue)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (f *FpcStubInterface) GetPublicState(key string) ([]byte, error) {
	value, err := f.stub.GetState(key)
	if err != nil {
		return nil, err
	}
	v_hash := sha256.Sum256(value)
	f.fpcKvSet.RwSet.Reads = append(f.fpcKvSet.RwSet.Reads, &kvrwset.KVRead{
		Key:     key,
		Version: nil,
	})
	f.fpcKvSet.ReadValueHashes = append(f.fpcKvSet.ReadValueHashes, v_hash[:])
	return value, nil
}

// PutState puts the specified `key` and `value` into the transaction's
// writeset as a data-write proposal. PutState doesn't effect the ledger
// until the transaction is validated and successfully committed.
// Simple keys must not be an empty string and must not start with a
// null character (0x00) in order to avoid range query collisions with
// composite keys, which internally get prefixed with 0x00 as composite
// key namespace. In addition, if using CouchDB, keys can only contain
// valid UTF-8 strings and cannot begin with an underscore ("_").
func (f *FpcStubInterface) PutState(key string, value []byte) error {
	encValue, err := f.csp.EncryptMessage(f.stateKey, value)
	if err != nil {
		return err
	}
	return f.PutPublicState(key, encValue)
}

func (f *FpcStubInterface) PutPublicState(key string, value []byte) error {
	f.fpcKvSet.RwSet.Writes = append(f.fpcKvSet.RwSet.Writes, &kvrwset.KVWrite{
		Key:      key,
		IsDelete: false,
		Value:    value,
	})
	return f.stub.PutState(key, value)
}

// DelState records the specified `key` to be deleted in the writeset of
// the transaction proposal. The `key` and its value will be deleted from
// the ledger when the transaction is validated and successfully committed.
func (f *FpcStubInterface) DelState(key string) error {
	panic("not implemented") // TODO: Implement
}

// SetStateValidationParameter sets the key-level endorsement policy for `key`.
func (f *FpcStubInterface) SetStateValidationParameter(key string, ep []byte) error {
	panic("not implemented") // TODO: Implement
}

// GetStateValidationParameter retrieves the key-level endorsement policy
// for `key`. Note that this will introduce a read dependency on `key` in
// the transaction's readset.
func (f *FpcStubInterface) GetStateValidationParameter(key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// GetStateByRange returns a range iterator over a set of keys in the
// ledger. The iterator can be used to iterate over all keys
// between the startKey (inclusive) and endKey (exclusive).
// However, if the number of keys between startKey and endKey is greater than the
// totalQueryLimit (defined in core.yaml), this iterator cannot be used
// to fetch all keys (results will be capped by the totalQueryLimit).
// The keys are returned by the iterator in lexical order. Note
// that startKey and endKey can be empty string, which implies unbounded range
// query on start or end.
// Call Close() on the returned StateQueryIteratorInterface object when done.
// The query is re-executed during validation phase to ensure result set
// has not changed since transaction endorsement (phantom reads detected).
func (f *FpcStubInterface) GetStateByRange(startKey string, endKey string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

// GetStateByRangeWithPagination returns a range iterator over a set of keys in the
// ledger. The iterator can be used to fetch keys between the startKey (inclusive)
// and endKey (exclusive).
// When an empty string is passed as a value to the bookmark argument, the returned
// iterator can be used to fetch the first `pageSize` keys between the startKey
// (inclusive) and endKey (exclusive).
// When the bookmark is a non-emptry string, the iterator can be used to fetch
// the first `pageSize` keys between the bookmark (inclusive) and endKey (exclusive).
// Note that only the bookmark present in a prior page of query results (ResponseMetadata)
// can be used as a value to the bookmark argument. Otherwise, an empty string must
// be passed as bookmark.
// The keys are returned by the iterator in lexical order. Note
// that startKey and endKey can be empty string, which implies unbounded range
// query on start or end.
// Call Close() on the returned StateQueryIteratorInterface object when done.
// This call is only supported in a read only transaction.
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

// GetStateByPartialCompositeKey queries the state in the ledger based on
// a given partial composite key. This function returns an iterator
// which can be used to iterate over all composite keys whose prefix matches
// the given partial composite key. However, if the number of matching composite
// keys is greater than the totalQueryLimit (defined in core.yaml), this iterator
// cannot be used to fetch all matching keys (results will be limited by the totalQueryLimit).
// The `objectType` and attributes are expected to have only valid utf8 strings and
// should not contain U+0000 (nil byte) and U+10FFFF (biggest and unallocated code point).
// See related functions SplitCompositeKey and CreateCompositeKey.
// Call Close() on the returned StateQueryIteratorInterface object when done.
// The query is re-executed during validation phase to ensure result set
// has not changed since transaction endorsement (phantom reads detected).
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
		decValue, err := f.csp.DecryptMessage(f.stateKey, i.Value)
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

// GetStateByPartialCompositeKeyWithPagination queries the state in the ledger based on
// a given partial composite key. This function returns an iterator
// which can be used to iterate over the composite keys whose
// prefix matches the given partial composite key.
// When an empty string is passed as a value to the bookmark argument, the returned
// iterator can be used to fetch the first `pageSize` composite keys whose prefix
// matches the given partial composite key.
// When the bookmark is a non-emptry string, the iterator can be used to fetch
// the first `pageSize` keys between the bookmark (inclusive) and the last matching
// composite key.
// Note that only the bookmark present in a prior page of query result (ResponseMetadata)
// can be used as a value to the bookmark argument. Otherwise, an empty string must
// be passed as bookmark.
// The `objectType` and attributes are expected to have only valid utf8 strings
// and should not contain U+0000 (nil byte) and U+10FFFF (biggest and unallocated
// code point). See related functions SplitCompositeKey and CreateCompositeKey.
// Call Close() on the returned StateQueryIteratorInterface object when done.
// This call is only supported in a read only transaction.
func (f *FpcStubInterface) GetStateByPartialCompositeKeyWithPagination(objectType string, keys []string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	panic("not implemented") // TODO: Implement
}

// CreateCompositeKey combines the given `attributes` to form a composite
// key. The objectType and attributes are expected to have only valid utf8
// strings and should not contain U+0000 (nil byte) and U+10FFFF
// (biggest and unallocated code point).
// The resulting composite key can be used as the key in PutState().
func (f *FpcStubInterface) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	key, err := f.stub.CreateCompositeKey(objectType, attributes)
	if err != nil {
		return "", err
	}
	return utils.TransformToFPCKey(key), nil
}

// SplitCompositeKey splits the specified key into attributes on which the
// composite key was formed. Composite keys found during range queries
// or partial composite key queries can therefore be split into their
// composite parts.
func (f *FpcStubInterface) SplitCompositeKey(compositeKey string) (string, []string, error) {
	panic("not implemented") // TODO: Implement
}

// GetQueryResult performs a "rich" query against a state database. It is
// only supported for state databases that support rich query,
// e.g.CouchDB. The query string is in the native syntax
// of the underlying state database. An iterator is returned
// which can be used to iterate over all keys in the query result set.
// However, if the number of keys in the query result set is greater than the
// totalQueryLimit (defined in core.yaml), this iterator cannot be used
// to fetch all keys in the query result set (results will be limited by
// the totalQueryLimit).
// The query is NOT re-executed during validation phase, phantom reads are
// not detected. That is, other committed transactions may have added,
// updated, or removed keys that impact the result set, and this would not
// be detected at validation/commit time.  Applications susceptible to this
// should therefore not use GetQueryResult as part of transactions that update
// ledger, and should limit use to read-only chaincode operations.
func (f *FpcStubInterface) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

// GetQueryResultWithPagination performs a "rich" query against a state database.
// It is only supported for state databases that support rich query,
// e.g., CouchDB. The query string is in the native syntax
// of the underlying state database. An iterator is returned
// which can be used to iterate over keys in the query result set.
// When an empty string is passed as a value to the bookmark argument, the returned
// iterator can be used to fetch the first `pageSize` of query results.
// When the bookmark is a non-emptry string, the iterator can be used to fetch
// the first `pageSize` keys between the bookmark and the last key in the query result.
// Note that only the bookmark present in a prior page of query results (ResponseMetadata)
// can be used as a value to the bookmark argument. Otherwise, an empty string
// must be passed as bookmark.
// This call is only supported in a read only transaction.
func (f *FpcStubInterface) GetQueryResultWithPagination(query string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	panic("not implemented") // TODO: Implement
}

// GetHistoryForKey returns a history of key values across time.
// For each historic key update, the historic value and associated
// transaction id and timestamp are returned. The timestamp is the
// timestamp provided by the client in the proposal header.
// GetHistoryForKey requires peer configuration
// core.ledger.history.enableHistoryDatabase to be true.
// The query is NOT re-executed during validation phase, phantom reads are
// not detected. That is, other committed transactions may have updated
// the key concurrently, impacting the result set, and this would not be
// detected at validation/commit time. Applications susceptible to this
// should therefore not use GetHistoryForKey as part of transactions that
// update ledger, and should limit use to read-only chaincode operations.
func (f *FpcStubInterface) GetHistoryForKey(key string) (shim.HistoryQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

// GetPrivateData returns the value of the specified `key` from the specified
// `collection`. Note that GetPrivateData doesn't read data from the
// private writeset, which has not been committed to the `collection`. In
// other words, GetPrivateData doesn't consider data modified by PutPrivateData
// that has not been committed.
func (f *FpcStubInterface) GetPrivateData(collection string, key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// GetPrivateDataHash returns the hash of the value of the specified `key` from the specified
// `collection`
func (f *FpcStubInterface) GetPrivateDataHash(collection string, key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// PutPrivateData puts the specified `key` and `value` into the transaction's
// private writeset. Note that only hash of the private writeset goes into the
// transaction proposal response (which is sent to the client who issued the
// transaction) and the actual private writeset gets temporarily stored in a
// transient store. PutPrivateData doesn't effect the `collection` until the
// transaction is validated and successfully committed. Simple keys must not
// be an empty string and must not start with a null character (0x00) in order
// to avoid range query collisions with composite keys, which internally get
// prefixed with 0x00 as composite key namespace. In addition, if using
// CouchDB, keys can only contain valid UTF-8 strings and cannot begin with an
// an underscore ("_").
func (f *FpcStubInterface) PutPrivateData(collection string, key string, value []byte) error {
	panic("not implemented") // TODO: Implement
}

// DelPrivateData records the specified `key` to be deleted in the private writeset
// of the transaction. Note that only hash of the private writeset goes into the
// transaction proposal response (which is sent to the client who issued the
// transaction) and the actual private writeset gets temporarily stored in a
// transient store. The `key` and its value will be deleted from the collection
// when the transaction is validated and successfully committed.
func (f *FpcStubInterface) DelPrivateData(collection string, key string) error {
	panic("not implemented") // TODO: Implement
}

// SetPrivateDataValidationParameter sets the key-level endorsement policy
// for the private data specified by `key`.
func (f *FpcStubInterface) SetPrivateDataValidationParameter(collection string, key string, ep []byte) error {
	panic("not implemented") // TODO: Implement
}

// GetPrivateDataValidationParameter retrieves the key-level endorsement
// policy for the private data specified by `key`. Note that this introduces
// a read dependency on `key` in the transaction's readset.
func (f *FpcStubInterface) GetPrivateDataValidationParameter(collection string, key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// GetPrivateDataByRange returns a range iterator over a set of keys in a
// given private collection. The iterator can be used to iterate over all keys
// between the startKey (inclusive) and endKey (exclusive).
// The keys are returned by the iterator in lexical order. Note
// that startKey and endKey can be empty string, which implies unbounded range
// query on start or end.
// Call Close() on the returned StateQueryIteratorInterface object when done.
// The query is re-executed during validation phase to ensure result set
// has not changed since transaction endorsement (phantom reads detected).
func (f *FpcStubInterface) GetPrivateDataByRange(collection string, startKey string, endKey string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

// GetPrivateDataByPartialCompositeKey queries the state in a given private
// collection based on a given partial composite key. This function returns
// an iterator which can be used to iterate over all composite keys whose prefix
// matches the given partial composite key. The `objectType` and attributes are
// expected to have only valid utf8 strings and should not contain
// U+0000 (nil byte) and U+10FFFF (biggest and unallocated code point).
// See related functions SplitCompositeKey and CreateCompositeKey.
// Call Close() on the returned StateQueryIteratorInterface object when done.
// The query is re-executed during validation phase to ensure result set
// has not changed since transaction endorsement (phantom reads detected).
func (f *FpcStubInterface) GetPrivateDataByPartialCompositeKey(collection string, objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

// GetPrivateDataQueryResult performs a "rich" query against a given private
// collection. It is only supported for state databases that support rich query,
// e.g.CouchDB. The query string is in the native syntax
// of the underlying state database. An iterator is returned
// which can be used to iterate (next) over the query result set.
// The query is NOT re-executed during validation phase, phantom reads are
// not detected. That is, other committed transactions may have added,
// updated, or removed keys that impact the result set, and this would not
// be detected at validation/commit time.  Applications susceptible to this
// should therefore not use GetPrivateDataQueryResult as part of transactions that update
// ledger, and should limit use to read-only chaincode operations.
func (f *FpcStubInterface) GetPrivateDataQueryResult(collection string, query string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented") // TODO: Implement
}

// GetCreator returns `SignatureHeader.Creator` (e.g. an identity)
// of the `SignedProposal`. This is the identity of the agent (or user)
// submitting the transaction.
func (f *FpcStubInterface) GetCreator() ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// GetTransient returns the `ChaincodeProposalPayload.Transient` field.
// It is a map that contains data (e.g. cryptographic material)
// that might be used to implement some form of application-level
// confidentiality. The contents of this field, as prescribed by
// `ChaincodeProposalPayload`, are supposed to always
// be omitted from the transaction and excluded from the ledger.
func (f *FpcStubInterface) GetTransient() (map[string][]byte, error) {
	panic("not implemented") // TODO: Implement
}

// GetBinding returns the transaction binding, which is used to enforce a
// link between application data (like those stored in the transient field
// above) to the proposal itself. This is useful to avoid possible replay
// attacks.
func (f *FpcStubInterface) GetBinding() ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// GetDecorations returns additional data (if applicable) about the proposal
// that originated from the peer. This data is set by the decorators of the
// peer, which append or mutate the chaincode input passed to the chaincode.
func (f *FpcStubInterface) GetDecorations() map[string][]byte {
	panic("not implemented") // TODO: Implement
}

// GetSignedProposal returns the SignedProposal object, which contains all
// data elements part of a transaction proposal.
func (f *FpcStubInterface) GetSignedProposal() (*pb.SignedProposal, error) {
	panic("not implemented") // TODO: Implement
}

// GetTxTimestamp returns the timestamp when the transaction was created. This
// is taken from the transaction ChannelHeader, therefore it will indicate the
// client's timestamp and will have the same value across all endorsers.
func (f *FpcStubInterface) GetTxTimestamp() (*timestamp.Timestamp, error) {
	panic("not implemented") // TODO: Implement
}

// SetEvent allows the chaincode to set an event on the response to the
// proposal to be included as part of a transaction. The event will be
// available within the transaction in the committed block regardless of the
// validity of the transaction.
// Only a single event can be included in a transaction, and must originate
// from the outer-most invoked chaincode in chaincode-to-chaincode scenarios.
// The marshaled ChaincodeEvent will be available in the transaction's ChaincodeAction.events field.
func (f *FpcStubInterface) SetEvent(name string, payload []byte) error {
	panic("not implemented") // TODO: Implement
}
