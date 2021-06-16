/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package crypto

// CSP (Crypto Service Provider) offers a high-level abstract of cryptographic primitives used in FPC
type CSP interface {
	NewRSAKeys() (publicKey []byte, privateKey []byte, e error)
	NewECDSAKeys() (publicKey []byte, privateKey []byte, e error)
	VerifyMessage(publicKey []byte, message []byte, signature []byte) error
	NewSymmetricKey() ([]byte, error)
	SignMessage(privateKey []byte, message []byte) (signature []byte, e error)
	PkDecryptMessage(privateKey []byte, encryptedMessage []byte) (message []byte, e error)
	PkEncryptMessage(publicKey []byte, message []byte) ([]byte, error)
	DecryptMessage(key []byte, encryptedMessage []byte) ([]byte, error)
	EncryptMessage(key []byte, message []byte) (encryptedMessage []byte, e error)
}

func GetDefaultCSP() CSP {
	return &GoCrypto{}
}
