/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCases struct {
	Name string
	CSP  CSP
}

var allTestCases = []testCases{
	// by default, we use the Go crypto implementation
	{"Go Crypto", NewGoCrypto()},
}

func TestNewRSAKeys(t *testing.T) {
	for _, tc := range allTestCases {
		pubKey, privKey, err := tc.CSP.NewRSAKeys()
		assert.NoError(t, err)

		assert.NotEmpty(t, pubKey)
		block, _ := pem.Decode(pubKey)
		assert.NotNil(t, block)
		assert.Equal(t, "RSA PUBLIC KEY", block.Type)
		pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
		assert.NotNil(t, pub)
		assert.NoError(t, err)

		assert.NotEmpty(t, privKey)
		block, _ = pem.Decode(privKey)
		assert.NotNil(t, block)
		assert.Equal(t, "RSA PRIVATE KEY", block.Type)
		pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		assert.NotNil(t, pri)
		assert.NoError(t, err)
	}
}

func TestNewECDSAKeys(t *testing.T) {
	for _, tc := range allTestCases {
		pubKey, privKey, err := tc.CSP.NewECDSAKeys()
		assert.NoError(t, err)

		assert.NotEmpty(t, pubKey)
		block, _ := pem.Decode(pubKey)
		assert.NotNil(t, block)
		assert.Equal(t, "PUBLIC KEY", block.Type)
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		assert.NotNil(t, pub)
		assert.NoError(t, err)

		assert.NotEmpty(t, privKey)
		block, _ = pem.Decode(privKey)
		assert.NotNil(t, block)
		assert.Equal(t, "EC PRIVATE KEY", block.Type)
		pri, err := x509.ParseECPrivateKey(block.Bytes)
		assert.NotNil(t, pri)
		assert.NoError(t, err)
	}
}

func TestNewSymmetricKey(t *testing.T) {
	for _, tc := range allTestCases {
		symKey, err := tc.CSP.NewSymmetricKey()
		assert.NotNil(t, symKey)
		assert.NoError(t, err)
	}
}

func TestSignature(t *testing.T) {
	msg := []byte("some message")

	for _, tc := range allTestCases {
		pubKey, privKey, err := tc.CSP.NewECDSAKeys()
		assert.NotEmpty(t, pubKey)
		assert.NotEmpty(t, privKey)
		assert.NoError(t, err)

		// fail with invalid key
		sig, err := tc.CSP.SignMessage([]byte("invalid key"), msg)
		assert.Nil(t, sig)
		assert.Error(t, err)

		// should succeed
		sig, err = tc.CSP.SignMessage(privKey, msg)
		assert.NoError(t, err)
		assert.NotNil(t, sig)

		err = tc.CSP.VerifyMessage([]byte("invalid key"), msg, sig)
		assert.Error(t, err)

		err = tc.CSP.VerifyMessage(pubKey, []byte("invalid msg"), sig)
		assert.Error(t, err)

		err = tc.CSP.VerifyMessage(pubKey, msg, []byte("invalid sig"))
		assert.Error(t, err)

		// should succeed
		err = tc.CSP.VerifyMessage(pubKey, msg, sig)
		assert.NoError(t, err)
	}
}

func TestPkEncryption(t *testing.T) {
	msg := []byte("some message")

	for _, tc := range allTestCases {
		pubKey, privKey, err := tc.CSP.NewRSAKeys()
		assert.NotEmpty(t, pubKey)
		assert.NotEmpty(t, privKey)
		assert.NoError(t, err)

		cipher, err := tc.CSP.PkEncryptMessage([]byte("invalid key"), msg)
		assert.Nil(t, cipher)
		assert.Error(t, err)

		// should succeed
		cipher, err = tc.CSP.PkEncryptMessage(pubKey, msg)
		assert.NotNil(t, cipher)
		assert.NoError(t, err)

		plain, err := tc.CSP.PkDecryptMessage([]byte("invalid key"), cipher)
		assert.Nil(t, plain)
		assert.Error(t, err)

		// should succeed
		plain, err = tc.CSP.PkDecryptMessage(privKey, cipher)
		assert.Equal(t, plain, msg)
		assert.NoError(t, err)
	}
}

func TestSymEncryption(t *testing.T) {
	msg := []byte("some message")

	for _, tc := range allTestCases {
		key, err := tc.CSP.NewSymmetricKey()
		assert.NotEmpty(t, key)
		assert.NoError(t, err)

		cipher, err := tc.CSP.EncryptMessage([]byte("invalid key"), msg)
		assert.Nil(t, cipher)
		assert.Error(t, err)

		// should succeed
		cipher, err = tc.CSP.EncryptMessage(key, msg)
		assert.NotNil(t, cipher)
		assert.NoError(t, err)

		plain, err := tc.CSP.DecryptMessage([]byte("invalid key"), cipher)
		assert.Nil(t, plain)
		assert.Error(t, err)

		// should succeed
		plain, err = tc.CSP.DecryptMessage(key, cipher)
		assert.Equal(t, plain, msg)
		assert.NoError(t, err)
	}
}
