/*
SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/internal/utils/fakes"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/assert"
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
