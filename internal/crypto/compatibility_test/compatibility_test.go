//go:build WITH_PDO_CRYPTO
// +build WITH_PDO_CRYPTO

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package compatibility_test

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name     string
	producer crypto.CSP
	verifier crypto.CSP
}

var allTestCases = []testCase{
	{"pdo -> go", crypto.NewPdoCrypto(), crypto.NewGoCrypto()},
	{"go -> pdo", crypto.NewGoCrypto(), crypto.NewPdoCrypto()},
}

func TestMixSignature(t *testing.T) {
	msg := []byte("some message")

	for _, tc := range allTestCases {
		fmt.Printf("run %s\n", tc.name)

		pubKey, privKey, err := tc.producer.NewECDSAKeys()
		assert.NotEmpty(t, pubKey)
		assert.NotEmpty(t, privKey)
		assert.NoError(t, err)

		// should succeed
		sig, err := tc.producer.SignMessage(privKey, msg)
		assert.NoError(t, err)
		assert.NotNil(t, sig)

		err = tc.verifier.VerifyMessage([]byte("invalid key"), msg, sig)
		assert.Error(t, err)

		err = tc.verifier.VerifyMessage(pubKey, []byte("invalid msg"), sig)
		assert.Error(t, err)

		err = tc.verifier.VerifyMessage(pubKey, msg, []byte("invalid sig"))
		assert.Error(t, err)

		// should succeed
		err = tc.verifier.VerifyMessage(pubKey, msg, sig)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}
}

func TestMixedPkEncryption(t *testing.T) {
	msg := []byte("some message")

	for _, tc := range allTestCases {
		fmt.Printf("run %s\n", tc.name)

		pubKey, privKey, err := tc.producer.NewRSAKeys()
		assert.NotEmpty(t, pubKey)
		assert.NotEmpty(t, privKey)
		assert.NoError(t, err)

		cipher, err := tc.producer.PkEncryptMessage([]byte("invalid key"), msg)
		assert.Nil(t, cipher)
		assert.Error(t, err)

		// should succeed
		cipher, err = tc.producer.PkEncryptMessage(pubKey, msg)
		assert.NotNil(t, cipher)
		assert.NoError(t, err)

		plain, err := tc.verifier.PkDecryptMessage([]byte("invalid key"), cipher)
		assert.Nil(t, plain)
		assert.Error(t, err)

		// should succeed
		plain, err = tc.verifier.PkDecryptMessage(privKey, cipher)
		assert.Equal(t, msg, plain)
		assert.NoError(t, err)
	}
}

func TestMixedSymEncryption(t *testing.T) {
	msg := []byte("some message")

	for _, tc := range allTestCases {
		fmt.Printf("run %s\n", tc.name)

		key, err := tc.producer.NewSymmetricKey()
		assert.NotEmpty(t, key)
		assert.NoError(t, err)

		cipher, err := tc.producer.EncryptMessage([]byte("invalid key"), msg)
		assert.Nil(t, cipher)
		assert.Error(t, err)

		// should succeed
		cipher, err = tc.producer.EncryptMessage(key, msg)
		assert.NotNil(t, cipher)
		assert.NoError(t, err)

		plain, err := tc.producer.DecryptMessage([]byte("invalid key"), cipher)
		assert.Nil(t, plain)
		assert.Error(t, err)

		// should succeed
		plain, err = tc.verifier.DecryptMessage(key, cipher)
		assert.Equal(t, msg, plain)
		assert.NoError(t, err)
	}
}
