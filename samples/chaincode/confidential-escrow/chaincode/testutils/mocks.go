package testutils

import (
	"container/list"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// MockStub provides a test implementation of MockStubInterface
// It simulates the Fabric ledger state for unit testing without requiring
// a running blockchain network.
type MockStub struct {
	State        map[string][]byte // State stores key-value pairs simulating the ledger
	TransientMap map[string][]byte // TransientMap stores transient data for the current transaction
	TxID         string            // TxID is the simulated transaction ID
	ChannelID    string            // ChannelID is the simulated channel name
	Creator      []byte            // Creator simulates the transaction creator's certificate
	Invocations  []string          // Invocations tracks function calls for verification
	Keys         *list.List
	// PropertyIndex map[string]map[string]string // assetType → property → key
}

// CompositeKeyRegistry defines which properties should be indexed for each asset type
// This makes the mock generic and extensible
var CompositeKeyRegistry = map[string][]string{
	"userdir":      {"publicKeyHash"},
	"wallet":       {"ownerPubKey"},
	"digitalAsset": {"symbol"},
	"escrow":       {"escrowId"},
}

// NewMockStub creates a new mock stub with initialized state
func NewMockStub() *MockStub {
	mockCert := `-----BEGIN CERTIFICATE-----
MIICJjCCAcygAwIBAgIQHv152Ul3TG/REl3mHfYyUjAKBggqhkjOPQQDAjBxMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEYMBYGA1UEChMPb3JnLmV4YW1wbGUuY29tMRswGQYDVQQDExJjYS5v
cmcuZXhhbXBsZS5jb20wHhcNMjQwNTA5MjEwOTAwWhcNMzQwNTA3MjEwOTAwWjBq
MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2Fu
IEZyYW5jaXNjbzEOMAwGA1UECxMFYWRtaW4xHjAcBgNVBAMMFUFkbWluQG9yZy5l
eGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABAoAt6mlUBMB0Ab1
paR0ILegN6qKmNfOYR0WV0kGOQwkO4lYcN76lSA2wSlWNTgxtGQDzja1708Ezdr5
vJ5KFhmjTTBLMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMCsGA1UdIwQk
MCKAID7vB1ct0j2yeNTm45AlCyj9TW22dYtjmPOGq+SVlMKQMAoGCCqGSM49BAMC
A0gAMEUCIQDYol2ylLCcz8qrGJmAFEG/cIG2Kxv8BD5t7Gv/28y8kgIgTz0Y75p6
3kbL5VN/PCiG2SbX72AVPSiEqj6PSiZJMz4=
-----END CERTIFICATE-----`

	// Create a mock serialized identity that cc-tools can parse
	mockIdentity := &msp.SerializedIdentity{
		Mspid:   "Org1MSP", // Match the Callers in your transactions
		IdBytes: []byte(mockCert),
	}

	creatorBytes, _ := proto.Marshal(mockIdentity)

	return &MockStub{
		State:        make(map[string][]byte),
		TransientMap: make(map[string][]byte),
		TxID:         "mock-tx-id",
		ChannelID:    "mock-channel",
		Creator:      creatorBytes,
		Invocations:  []string{},
		Keys:         list.New(),
	}
}

// GetState retrieves the value for a given key from mock state
func (m *MockStub) GetState(key string) ([]byte, error) {
	fmt.Printf("DEBUG GetState called with key: %q\n", key)
	m.Invocations = append(m.Invocations, fmt.Sprintf("GetState:%s", key))

	// First: Always try direct lookup
	if value, exists := m.State[key]; exists {
		fmt.Printf("DEBUG Direct lookup SUCCESS for key: %q\n", key)
		return value, nil
	}
	fmt.Printf("DEBUG Direct lookup FAILED for key: %q\n", key)

	// Second: If key matches "assetType:uuid" pattern, search composite keys
	if strings.Contains(key, ":") {
		parts := strings.Split(key, ":")
		if len(parts) == 2 {
			assetType := parts[0]
			fmt.Printf("DEBUG Searching composite keys for assetType: %q\n", assetType)

			// Search all composite keys for this asset type
			compositePrefix := assetType + "\x00"
			fmt.Printf("DEBUG Composite prefix: %q\n", compositePrefix)

			// Just return the FIRST matching composite key
			// Don't try to match @key since CCTools generates different UUIDs
			for stateKey, value := range m.State {
				if strings.HasPrefix(stateKey, compositePrefix) {
					fmt.Printf("DEBUG Found composite key: %q - RETURNING IT\n", stateKey)
					return value, nil
				}
			}

			fmt.Printf("DEBUG No composite keys found with prefix: %q\n", compositePrefix)
		}
	}

	fmt.Printf("DEBUG GetState returning nil for key: %q\n", key)
	return nil, nil
}

// PutState stores a key-value pair in mock state
func (m *MockStub) PutState(key string, value []byte) error {
	fmt.Printf("DEBUG PutState called with key: %q\n", key)
	m.Invocations = append(m.Invocations, fmt.Sprintf("PutState:%s", key))

	if len(value) == 0 {
		delete(m.State, key)
		return nil
	}

	m.State[key] = value
	fmt.Printf("DEBUG Stored key: %q\n", key)

	// Generic composite key indexing
	var assetData map[string]any
	if err := json.Unmarshal(value, &assetData); err == nil {
		assetType, _ := assetData["@assetType"].(string)
		fmt.Printf("DEBUG Asset type: %q\n", assetType)

		// Check if this asset type has registered composite key properties
		if indexedProps, exists := CompositeKeyRegistry[assetType]; exists {
			fmt.Printf("DEBUG Indexed properties for %q: %v\n", assetType, indexedProps)

			// For each indexed property, create a composite key
			for _, propName := range indexedProps {
				if propValue, ok := assetData[propName].(string); ok && propValue != "" {
					compositeKey := assetType + string('\x00') + propValue
					m.State[compositeKey] = value
					fmt.Printf("DEBUG Created composite key: %q (property: %s=%s)\n", compositeKey, propName, propValue)
				} else {
					fmt.Printf("DEBUG Property %q not found or empty\n", propName)
				}
			}
		} else {
			fmt.Printf("DEBUG No composite key registry entry for asset type: %q\n", assetType)
		}
	}

	// Maintain ordered key list
	inserted := false
	for elem := m.Keys.Front(); elem != nil; elem = elem.Next() {
		elemValue := elem.Value.(string)
		comp := strings.Compare(key, elemValue)
		if comp < 0 {
			m.Keys.InsertBefore(key, elem)
			inserted = true
			break
		} else if comp == 0 {
			inserted = true
			break
		}
	}

	if !inserted {
		if m.Keys.Len() == 0 {
			m.Keys.PushFront(key)
		} else {
			m.Keys.PushBack(key)
		}
	}

	return nil
}

// func (m *MockStub) PutState(key string, value []byte) error {
// 	m.Invocations = append(m.Invocations, fmt.Sprintf("PutState:%s", key))
//
// 	// If value is empty, delete the key
// 	if len(value) == 0 {
// 		delete(m.State, key)
// 		return nil
// 	}
//
// 	m.State[key] = value
//
// 	// Generic composite key indexing for any registered asset type
// 	var assetData map[string]any
// 	if err := json.Unmarshal(value, &assetData); err == nil {
// 		assetType, _ := assetData["@assetType"].(string)
//
// 		// Check if this asset type has registered composite key properties
// 		if indexedProps, exists := CompositeKeyRegistry[assetType]; exists {
// 			// For each indexed property, create a composite key
// 			for _, propName := range indexedProps {
// 				if propValue, ok := assetData[propName].(string); ok && propValue != "" {
// 					// Create composite key: assetType\x00propertyValue
// 					compositeKey := assetType + string('\x00') + propValue
// 					m.State[compositeKey] = value
// 				}
// 			}
// 		}
// 	}
//
// 	// Maintain ordered key list
// 	inserted := false
// 	for elem := m.Keys.Front(); elem != nil; elem = elem.Next() {
// 		elemValue := elem.Value.(string)
// 		comp := strings.Compare(key, elemValue)
// 		if comp < 0 {
// 			m.Keys.InsertBefore(key, elem)
// 			inserted = true
// 			break
// 		} else if comp == 0 {
// 			// Key already exists
// 			inserted = true
// 			break
// 		}
// 	}
//
// 	// If not inserted and list is not empty, add to end
// 	if !inserted {
// 		if m.Keys.Len() == 0 {
// 			m.Keys.PushFront(key)
// 		} else {
// 			m.Keys.PushBack(key)
// 		}
// 	}
//
// 	return nil
// }

// DelState removes a key from mock state
func (m *MockStub) DelState(key string) error {
	m.Invocations = append(m.Invocations, fmt.Sprintf("DeleteState:%s", key))
	delete(m.State, key)
	return nil
}

// GetStateByRange returns an iterator for keys within a range
func (m *MockStub) GetStateByRange(startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	return NewMockStateRangeQueryIterator(m, startKey, endKey), nil
}

// GetStateByPartialCompositeKey returns an iterator for composite keys
func (m *MockStub) GetStateByPartialCompositeKey(objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	partialCompositeKey := objectType
	for _, key := range keys {
		partialCompositeKey += string('\x00') + key
	}

	fmt.Printf("DEBUG GetStateByPartialCompositeKey called:\n")
	fmt.Printf("DEBUG   objectType: %q\n", objectType)
	fmt.Printf("DEBUG   keys: %v\n", keys)
	fmt.Printf("DEBUG   partialCompositeKey: %q\n", partialCompositeKey)

	fmt.Printf("DEBUG Available keys in state:\n")
	for k := range m.State {
		fmt.Printf("DEBUG   %q\n", k)
	}

	// Use range query from partial key to partial key + max unicode
	return NewMockStateRangeQueryIterator(m, partialCompositeKey, partialCompositeKey+string(rune(0x10FFFF))), nil
}

// GetQueryResult executes a rich query (not implemented in mock)
func (m *MockStub) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	return NewMockStateRangeQueryIterator(m, "", ""), nil
}

// GetHistoryForKey returns history for a key (not implemented in mock)
func (m *MockStub) GetHistoryForKey(key string) (shim.HistoryQueryIteratorInterface, error) {
	return &MockHistoryIterator{}, nil
}

// CreateCompositeKey creates a composite key
func (m *MockStub) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	// return objectType + ":" + attributes[0], nil
	key := objectType
	for _, attr := range attributes {
		key += string('\x00') + attr
	}
	// fmt.Printf("DEBUG CreateCompositeKey: objectType=%q, attributes=%v, result=%q\n", objectType, attributes, key)
	return key, nil
}

// SplitCompositeKey splits a composite key
func (m *MockStub) SplitCompositeKey(compositeKey string) (string, []string, error) {
	return "", []string{}, nil
}

// GetTransient returns the transient map
func (m *MockStub) GetTransient() (map[string][]byte, error) {
	return m.TransientMap, nil
}

// GetTxID returns the transaction ID
func (m *MockStub) GetTxID() string {
	return m.TxID
}

// GetChannelID returns the channel ID
func (m *MockStub) GetChannelID() string {
	return m.ChannelID
}

// GetTxTimestamp returns a mock timestamp
func (m *MockStub) GetTxTimestamp() (*timestamp.Timestamp, error) {
	now := time.Now()
	return &timestamp.Timestamp{
		Seconds: now.Unix(),
		Nanos:   int32(now.Nanosecond()),
	}, nil
}

// GetCreator returns the transaction creator
func (m *MockStub) GetCreator() ([]byte, error) {
	return m.Creator, nil
}

// GetDecorations is a no-op
func (s *MockStub) GetDecorations() map[string][]byte {
	return make(map[string][]byte)
}

// GetBinding returns empty binding
func (m *MockStub) GetBinding() ([]byte, error) {
	return []byte{}, nil
}

// GetSignedProposal returns nil
func (m *MockStub) GetSignedProposal() (*peer.SignedProposal, error) {
	return nil, nil
}

// GetArgs returns empty args
func (m *MockStub) GetArgs() [][]byte {
	return [][]byte{}
}

// GetStringArgs returns empty args
func (m *MockStub) GetStringArgs() []string {
	return []string{}
}

// GetFunctionAndParameters returns mock function name
func (m *MockStub) GetFunctionAndParameters() (string, []string) {
	return "mockFunction", []string{}
}

// GetArgsSlice returns empty slice
func (m *MockStub) GetArgsSlice() ([]byte, error) {
	return []byte{}, nil
}

// SetEvent sets an event (no-op in mock)
func (m *MockStub) SetEvent(name string, payload []byte) error {
	return nil
}

// InvokeChaincode simulates chaincode invocation (not implemented)
func (m *MockStub) InvokeChaincode(chaincodeName string, args [][]byte, channel string) peer.Response {
	return shim.Success(nil)
}

// GetStateValidationParameter returns nil
func (m *MockStub) GetStateValidationParameter(key string) ([]byte, error) {
	return nil, nil
}

// SetStateValidationParameter is a no-op
func (m *MockStub) SetStateValidationParameter(key string, ep []byte) error {
	return nil
}

// GetPrivateData returns nil
func (m *MockStub) GetPrivateData(collection, key string) ([]byte, error) {
	return nil, nil
}

// GetPrivateDataHash is a no-op
func (s *MockStub) GetPrivateDataHash(collection string, key string) ([]byte, error) {
	return nil, nil
}

// PutPrivateData is a no-op
func (m *MockStub) PutPrivateData(collection, key string, value []byte) error {
	return nil
}

// DelPrivateData is a no-op
func (m *MockStub) DelPrivateData(collection, key string) error {
	return nil
}

// GetPrivateDataByRange returns empty iterator
func (m *MockStub) GetPrivateDataByRange(collection, startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	return NewMockStateRangeQueryIterator(m, startKey, endKey), nil
}

// GetPrivateDataByPartialCompositeKey returns empty iterator
func (m *MockStub) GetPrivateDataByPartialCompositeKey(collection, objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	return NewMockStateRangeQueryIterator(m, "", ""), nil
}

// GetPrivateDataQueryResult returns empty iterator
func (m *MockStub) GetPrivateDataQueryResult(collection, query string) (shim.StateQueryIteratorInterface, error) {
	return NewMockStateRangeQueryIterator(m, "", ""), nil
}

// GetPrivateDataValidationParameter returns nil
func (m *MockStub) GetPrivateDataValidationParameter(collection, key string) ([]byte, error) {
	return nil, nil
}

// SetPrivateDataValidationParameter is a no-op
func (m *MockStub) SetPrivateDataValidationParameter(collection, key string, ep []byte) error {
	return nil
}

// GetQueryResultWithPagination is a no-op
func (s *MockStub) GetQueryResultWithPagination(query string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, nil
}

// GetStateByPartialCompositeKeyWithPagination is a no-op
func (s *MockStub) GetStateByPartialCompositeKeyWithPagination(objectType string, keys []string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, nil
}

// GetStateByRangeWithPagination is a no-op
func (s *MockStub) GetStateByRangeWithPagination(startKey, endKey string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, nil
}

// PurgePrivateData is a no-op
func (s *MockStub) PurgePrivateData(collection string, key string) error {
	return nil
}

// ///////////////////////////////////////////////////////////////
// MockHistoryIterator implements HistoryQueryIteratorInterface //
// ///////////////////////////////////////////////////////////////
type MockHistoryIterator struct{}

func (m *MockHistoryIterator) HasNext() bool {
	return false
}

func (m *MockHistoryIterator) Next() (*queryresult.KeyModification, error) {
	return nil, fmt.Errorf("no history")
}

func (m *MockHistoryIterator) Close() error {
	return nil
}

// ////////////////////////////////////////////////////////////
// MockStubWrapper wraps MockStub for cc-tools compatibility //
// ////////////////////////////////////////////////////////////
type MockStubWrapper struct {
	*sw.StubWrapper
	mockStub *MockStub
}

// NewMockStubWrapper creates a wrapped mock stub
func NewMockStubWrapper() (*MockStubWrapper, *MockStub) {
	mockStub := NewMockStub()
	wrapper := &sw.StubWrapper{
		Stub: mockStub,
	}
	return &MockStubWrapper{
		StubWrapper: wrapper,
		mockStub:    mockStub,
	}, mockStub
}

// GetMockStub returns the underlying mock stub for assertions
func (m *MockStubWrapper) GetMockStub() *MockStub {
	return m.mockStub
}

// //////////////////////////////
// MockStateRangeQueryIterator //
// //////////////////////////////
type MockStateRangeQueryIterator struct {
	Closed   bool
	Stub     *MockStub
	StartKey string
	EndKey   string
	Current  *list.Element
}

func (iter *MockStateRangeQueryIterator) HasNext() bool {
	if iter.Closed {
		return false
	}
	if iter.Current == nil {
		return false
	}

	current := iter.Current
	for current != nil {
		if iter.StartKey == "" && iter.EndKey == "" {
			return true
		}
		key := current.Value.(string)
		if strings.Compare(key, iter.StartKey) >= 0 && strings.Compare(key, iter.EndKey) < 0 {
			return true
		}
		if strings.Compare(key, iter.EndKey) >= 0 {
			return false
		}
		current = current.Next()
	}
	return false
}

func (iter *MockStateRangeQueryIterator) Next() (*queryresult.KV, error) {
	if iter.Closed {
		return nil, fmt.Errorf("iterator closed")
	}
	if !iter.HasNext() {
		return nil, fmt.Errorf("no more elements")
	}

	for iter.Current != nil {
		key := iter.Current.Value.(string)
		if strings.Compare(key, iter.StartKey) >= 0 && strings.Compare(key, iter.EndKey) < 0 {
			value := iter.Stub.State[key]
			iter.Current = iter.Current.Next()
			return &queryresult.KV{Key: key, Value: value}, nil
		}
		iter.Current = iter.Current.Next()
	}
	return nil, fmt.Errorf("no matching key found")
}

func (iter *MockStateRangeQueryIterator) Close() error {
	iter.Closed = true
	return nil
}

func NewMockStateRangeQueryIterator(stub *MockStub, startKey, endKey string) *MockStateRangeQueryIterator {
	return &MockStateRangeQueryIterator{
		Closed:   false,
		Stub:     stub,
		StartKey: startKey,
		EndKey:   endKey,
		Current:  stub.Keys.Front(),
	}
}
