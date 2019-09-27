/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/ercc"
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
)

const enclaveLibFile = "lib/enclave.signed.so"

func TestEnclaveStub_RemoteAttestation(t *testing.T) {
	stub := NewEnclave()
	err := stub.Create(enclaveLibFile)
	if err != nil {
		t.Fatalf("Create returned error %s", err)
	}

	// start without binding
	stub.Bind(nil, nil)

	mockRegistry := ercc.MockEnclaveRegistryStub{}
	spid, _ := mockRegistry.GetSPID(nil, "", "")

	//TODO: retrieve sigrl
	sig_rl := []byte(nil)
	sig_rl_size := uint(0)

	quoteAsBytes, pkBytes, err := stub.GetRemoteAttestationReport(spid, sig_rl, sig_rl_size)
	if err != nil {
		t.Fatalf("Attestation returned error %s", err)
	}
	if quoteAsBytes == nil {
		t.Fatalf("quote is nil")
	}
	if pkBytes == nil {
		t.Fatalf("pkBytes is nil")
	}

	// check pk is in quote
	q, err := attestation.QuoteFromBytes(quoteAsBytes)
	if err != nil {
		t.Fatalf("Can not parse quote %s", err)
	}

	t.Logf("pk der:\n%s", hex.Dump(pkBytes))

	pub, err := x509.ParsePKIXPublicKey(pkBytes)
	if err != nil {
		t.Fatalf("x509.ParsePKIXPublicKey error %s", err)
	}

	ecdsaPublickey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("enclave key is not ecdsa key")
	}

	t.Logf("X:\n%s\nY:\n%s", hex.Dump(ecdsaPublickey.X.Bytes()), hex.Dump(ecdsaPublickey.Y.Bytes()))

	// calc pk raw hash which is in quote
	h := sha256.New()
	h.Write(ecdsaPublickey.X.Bytes())
	h.Write(ecdsaPublickey.Y.Bytes())
	enclavePKRawHash := h.Sum(nil)

	if !reflect.DeepEqual(enclavePKRawHash[:32], q.ReportData[:32]) {
		t.Fatalf("enclave pk does not match quote")
	}

	enclavePkHash := sha256.Sum256(pkBytes)

	t.Logf("Enclave pk base64:\n%s", base64.StdEncoding.EncodeToString(pkBytes))
	t.Logf("Enclave pk hash base64:\n%s", base64.StdEncoding.EncodeToString(enclavePkHash[:]))
	t.Logf("Quote base64:\n%s", base64.StdEncoding.EncodeToString(quoteAsBytes))
}

func TestEnclaveStub_Invoke(t *testing.T) {
	stub := NewEnclave()
	err := stub.Create(enclaveLibFile)
	if err != nil {
		t.Fatalf("Create returned error %s", err)
	}

	// start without binding
	stub.Bind(nil, nil)

	_, _, err = stub.Invoke(nil, nil, nil, nil)
	if err == nil {
		t.Fatalf("error expected")
	}
}

func TestEnclaveStub_GetPublicKey(t *testing.T) {
	stub := NewEnclave()
	err := stub.Create(enclaveLibFile)
	if err != nil {
		t.Fatalf("Create returned error %s", err)
	}

	// start without binding
	stub.Bind(nil, nil)

	pk, err := stub.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPubKey returned error %s", err)
	}
	if pk == nil {
		t.Fatalf("pk is nil")
	}
}

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
