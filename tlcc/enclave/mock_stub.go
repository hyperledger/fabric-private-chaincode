/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
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
