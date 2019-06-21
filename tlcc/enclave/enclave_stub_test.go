/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
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
