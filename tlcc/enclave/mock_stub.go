/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave

// MockStub implements the interface
type MockStub struct {
}

func (m *MockStub) GetTargetInfo() ([]byte, error) {
	return []byte{}, nil
}

// Return report and enclave PK in DER-encoded PKIX format
func (m *MockStub) GetLocalAttestationReport(targetInfo []byte) ([]byte, []byte, error) {
	return []byte{}, []byte{}, nil
}

// Creates an enclave from a given enclave lib file
func (m *MockStub) Create(enclaveLibFile string) error {
	return nil
}

// Init enclave with a given genesis block
func (m *MockStub) InitWithGenesis(blockBytes []byte) error {
	return nil
}

// give enclave next block to validate and append to the ledger
func (m *MockStub) NextBlock(blockBytes []byte) error {
	return nil
}

// verifies state and returns cmac
func (m *MockStub) GetStateMetadata(key string, nonce []byte, isRangeQuery bool) ([]byte, error) {
	return []byte{}, nil
}

// Destroys enclave
func (m *MockStub) Destroy() error {
	return nil
}
