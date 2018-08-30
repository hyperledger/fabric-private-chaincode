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

func (m *MockStub) Attestation(spid []byte) ([]byte, []byte, error) {
	return []byte{}, []byte{}, nil
}

func (m *MockStub) Invoke(args []byte, input []byte) ([]byte, []byte, []byte, error) {
	return []byte{}, []byte{}, []byte{}, nil
}

func (m *MockStub) GetPublicKey() ([]byte, error) {
	return []byte{}, nil
}

func (m *MockStub) Create(enclaveLibFile string) error {
	return nil
}
func (m *MockStub) Destroy() error {
	return nil
}
