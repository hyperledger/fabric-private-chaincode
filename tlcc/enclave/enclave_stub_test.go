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

import (
	"testing"
)

const enclaveLibFile = "lib/enclave.signed.so"

func TestEnclaveStub_Create(t *testing.T) {
	stub := NewEnclave()
	err := stub.Create(enclaveLibFile)
	if err != nil {
		t.Fatalf("Create returned error %s", err)
	}
}

func TestEnclaveStub_Destroy(t *testing.T) {
	stub := NewEnclave()
	err := stub.Create(enclaveLibFile)
	if err != nil {
		t.Fatalf("Create returned error %s", err)
	}
	err = stub.Destroy()
	if err != nil {
		t.Fatalf("Deate returned error %s", err)
	}
}
