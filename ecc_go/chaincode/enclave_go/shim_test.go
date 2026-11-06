/*
SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/internal/endorsement"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils/fakes"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type plaintextState struct{}

func (plaintextState) EncryptState(plaintext []byte) ([]byte, error) { return plaintext, nil }

func (plaintextState) DecryptState(ciphertext []byte) ([]byte, error) { return ciphertext, nil }

func newTestStub(ledger map[string][]byte) *fakes.ChaincodeStub {
	stub := &fakes.ChaincodeStub{}
	stub.CreateCompositeKeyStub = func(objectType string, attributes []string) (string, error) {
		key := "\x00" + objectType + "\x00"
		for _, a := range attributes {
			key += a + "\x00"
		}
		return key, nil
	}
	stub.GetStateStub = func(key string) ([]byte, error) {
		return ledger[key], nil
	}
	stub.PutStateStub = func(key string, value []byte) error {
		ledger[key] = value
		return nil
	}
	stub.DelStateStub = func(key string) error {
		delete(ledger, key)
		return nil
	}
	return stub
}

func TestGetStateWithCompositeKey(t *testing.T) {
	ledger := map[string][]byte{
		"\x00asset\x00a1\x00": []byte("value"),
		"plain":               []byte("plain value"),
	}
	rwset := NewReadWriteSet()
	f := NewFpcStubInterface(newTestStub(ledger), &pb.ChaincodeInput{}, rwset, plaintextState{})

	key, err := f.CreateCompositeKey("asset", []string{"a1"})
	assert.NoError(t, err)
	assert.Equal(t, ".asset.a1.", key)

	value, err := f.GetState(key)
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), value)
	assert.Equal(t, hash([]byte("value")), rwset.reads[key].hash)

	value, err = f.GetPublicState(key)
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), value)

	value, err = f.GetState("plain")
	assert.NoError(t, err)
	assert.Equal(t, []byte("plain value"), value)
}

func TestGetStateWithDotKey(t *testing.T) {
	ledger := map[string][]byte{
		".": []byte("dot value"),
	}
	rwset := NewReadWriteSet()
	f := NewFpcStubInterface(newTestStub(ledger), &pb.ChaincodeInput{}, rwset, plaintextState{})

	value, err := f.GetState(".")
	assert.NoError(t, err)
	assert.Equal(t, []byte("dot value"), value)
}

// The shim records FPC keys in the rwset and the validator converts them
// to Fabric keys when it replays the rwset, so both sides must agree on
// the ledger key of a composite key.
func TestCompositeKeyRoundTripThroughValidator(t *testing.T) {
	ledger := map[string][]byte{}
	stub := newTestStub(ledger)
	validator := endorsement.NewValidator()

	// write a composite key and commit it
	rwset := NewReadWriteSet()
	f := NewFpcStubInterface(stub, &pb.ChaincodeInput{}, rwset, plaintextState{})
	key, err := f.CreateCompositeKey("asset", []string{"a1"})
	require.NoError(t, err)
	require.NoError(t, f.PutState(key, []byte("value")))
	require.NoError(t, validator.ReplayReadWrites(stub, rwset.ToFPCKVSet()))
	assert.Equal(t, map[string][]byte{"\x00asset\x00a1\x00": []byte("value")}, ledger)

	// read it back in a new transaction and replay the read
	rwset = NewReadWriteSet()
	f = NewFpcStubInterface(stub, &pb.ChaincodeInput{}, rwset, plaintextState{})
	value, err := f.GetState(key)
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), value)
	require.NoError(t, validator.ReplayReadWrites(stub, rwset.ToFPCKVSet()))

	// delete it in a third transaction
	rwset = NewReadWriteSet()
	f = NewFpcStubInterface(stub, &pb.ChaincodeInput{}, rwset, plaintextState{})
	require.NoError(t, f.DelState(key))
	require.NoError(t, validator.ReplayReadWrites(stub, rwset.ToFPCKVSet()))
	assert.Empty(t, ledger)
}
