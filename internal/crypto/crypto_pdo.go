//go:build WITH_PDO_CRYPTO
// +build WITH_PDO_CRYPTO

/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"crypto/rand"
	"fmt"
)

// #cgo CFLAGS: -I${SRCDIR}/../../common/crypto
// #cgo LDFLAGS: -L${SRCDIR}/../../common/crypto/_build -L${SRCDIR}/../../common/logging/_build -Wl,--start-group -lupdo-crypto-adapt -lupdo-crypto -Wl,--end-group -lcrypto -lulogging -lstdc++ -lgcov
// #include <stdio.h>
// #include <stdlib.h>
// #include <stdbool.h>
// #include <stdint.h>
// #include "pdo-crypto-c-wrapper.h"
import "C"

// PDOCrypto implements CSP using the PDO crypto library in common/crypto
type PDOCrypto struct {
}

func NewPdoCrypto() *PDOCrypto {
	return &PDOCrypto{}
}

// NewRSAKeys generates a new public/private RSA key pair
// The returned RSA keys created by the PDO crypto lib are 3072bit long and are PEM encoded
func (c PDOCrypto) NewRSAKeys() (publicKey []byte, privateKey []byte, e error) {
	//Here we roughly estimate that they fit in 3KB
	const estimatedPemRsaLen = 3072
	const serializedPublicKeyLen = estimatedPemRsaLen
	serializedPublicKeyPtr := C.malloc(serializedPublicKeyLen)
	defer C.free(serializedPublicKeyPtr)
	serializedPublicKeyActualLen := C.uint32_t(0)

	const serializedPrivateKeyLen = estimatedPemRsaLen
	serializedPrivateKeyPtr := C.malloc(serializedPrivateKeyLen)
	defer C.free(serializedPrivateKeyPtr)
	serializedPrivateKeyActualLen := C.uint32_t(0)

	ret := C.new_rsa_key(
		(*C.uint8_t)(serializedPublicKeyPtr),
		serializedPublicKeyLen,
		&serializedPublicKeyActualLen,
		(*C.uint8_t)(serializedPrivateKeyPtr),
		serializedPrivateKeyLen,
		&serializedPrivateKeyActualLen,
	)
	if !ret {
		return nil, nil, fmt.Errorf("cannot create RSA keys")
	}

	return C.GoBytes(serializedPublicKeyPtr, C.int(serializedPublicKeyActualLen)), C.GoBytes(serializedPrivateKeyPtr, C.int(serializedPrivateKeyActualLen)), nil
}

// NewECDSAKeys generates a new public/private ECDSA key pair
// The returned ECDSA keys are created by the PDO crypto lib and are PEM encoded
func (c PDOCrypto) NewECDSAKeys() (publicKey []byte, privateKey []byte, e error) {
	//Here we roughly estimate that they fit in 2KB
	const estimatedPemEcdsaLen = 2048
	const serializedPublicKeyLen = estimatedPemEcdsaLen
	serializedPublicKeyPtr := C.malloc(serializedPublicKeyLen)
	defer C.free(serializedPublicKeyPtr)
	serializedPublicKeyActualLen := C.uint32_t(0)

	const serializedPrivateKeyLen = estimatedPemEcdsaLen
	serializedPrivateKeyPtr := C.malloc(serializedPrivateKeyLen)
	defer C.free(serializedPrivateKeyPtr)
	serializedPrivateKeyActualLen := C.uint32_t(0)

	ret := C.new_ecdsa_key(
		(*C.uint8_t)(serializedPublicKeyPtr),
		serializedPublicKeyLen,
		&serializedPublicKeyActualLen,
		(*C.uint8_t)(serializedPrivateKeyPtr),
		serializedPrivateKeyLen,
		&serializedPrivateKeyActualLen,
	)
	if !ret {
		return nil, nil, fmt.Errorf("cannot create ECDSA keys")
	}

	return C.GoBytes(serializedPublicKeyPtr, C.int(serializedPublicKeyActualLen)), C.GoBytes(serializedPrivateKeyPtr, C.int(serializedPrivateKeyActualLen)), nil
}

// NewSymmetricKey generates a new symmetric key with the specified key length is required by the pdo crypto library
func (c PDOCrypto) NewSymmetricKey() ([]byte, error) {
	keyLength := C.SYM_KEY_LEN
	key := make([]byte, keyLength)
	n, err := rand.Read(key)
	if n != len(key) || err != nil {
		return nil, err
	}

	return key, nil
}

func (c PDOCrypto) VerifyMessage(publicKey []byte, message []byte, signature []byte) error {

	publicKeyPtr := C.CBytes(publicKey)
	defer C.free(publicKeyPtr)

	messagePtr := C.CBytes(message)
	defer C.free(messagePtr)

	signaturePtr := C.CBytes(signature)
	defer C.free(signaturePtr)

	ret := C.verify_signature(
		(*C.uint8_t)(publicKeyPtr),
		(C.uint32_t)(len(publicKey)),
		(*C.uint8_t)(messagePtr),
		(C.uint32_t)(len(message)),
		(*C.uint8_t)(signaturePtr),
		(C.uint32_t)(len(signature)))

	if !ret {
		return fmt.Errorf("verification failed")
	}

	return nil
}

func (c PDOCrypto) SignMessage(privateKey []byte, message []byte) (signature []byte, e error) {
	privateKeyPtr := C.CBytes(privateKey)
	defer C.free(privateKeyPtr)

	messagePtr := C.CBytes(message)
	defer C.free(messagePtr)

	// TODO why are we using RSA_KEY_SIZE here? sign_message uses ecdsa
	estimatedSignatureLen := C.RSA_KEY_SIZE >> 3 //bits-to-bytes conversion
	signaturePtr := C.malloc(C.ulong(estimatedSignatureLen))
	defer C.free(signaturePtr)
	signatureActualLen := C.uint32_t(0)

	ret := C.sign_message(
		(*C.uint8_t)(privateKeyPtr),
		(C.uint32_t)(len(privateKey)),
		(*C.uint8_t)(messagePtr),
		(C.uint32_t)(len(message)),
		(*C.uint8_t)(signaturePtr),
		(C.uint32_t)(estimatedSignatureLen),
		&signatureActualLen)
	if !ret {
		return nil, fmt.Errorf("cannot sign message")
	}

	return C.GoBytes(signaturePtr, C.int(signatureActualLen)), nil
}

// PkDecryptMessage is an RSA decryption performed with the pdo crypto library
// Importantly, the library uses 3072bit RSA keys & OAEP encoding, so the input size can be at most ~300bytes
func (c PDOCrypto) PkDecryptMessage(privateKey []byte, encryptedMessage []byte) (message []byte, e error) {

	privateKeyPtr := C.CBytes(privateKey)
	defer C.free(privateKeyPtr)

	encryptedMessagePtr := C.CBytes(encryptedMessage)
	defer C.free(encryptedMessagePtr)

	//estimate that the decrypted message will not be larger than the encrypted one
	decryptedMessageLen := len(encryptedMessage)
	decryptedMessagePtr := C.malloc(C.ulong(decryptedMessageLen))
	defer C.free(decryptedMessagePtr)

	decryptedMessageActualLen := C.uint32_t(0)

	ret := C.pk_decrypt_message(
		(*C.uint8_t)(privateKeyPtr),
		(C.uint32_t)(len(privateKey)),
		(*C.uint8_t)(encryptedMessagePtr),
		(C.uint32_t)(len(encryptedMessage)),
		(*C.uint8_t)(decryptedMessagePtr),
		(C.uint32_t)(decryptedMessageLen),
		&decryptedMessageActualLen,
	)
	if !ret {
		return nil, fmt.Errorf("pk decryption failed")
	}

	return C.GoBytes(decryptedMessagePtr, C.int(decryptedMessageActualLen)), nil
}

// PkEncryptMessage is an RSA encryption performed with the pdo crypto library
// It requires an RSA public key of size RSA_KEY_SIZE as defined in pdo-crypto-c-wrapper.h
// Importantly, the library uses 3072bit RSA keys & OAEP encoding, so the input size can be at most ~300bytes
func (c PDOCrypto) PkEncryptMessage(publicKey []byte, message []byte) ([]byte, error) {

	messagePtr := C.CBytes(message)
	defer C.free(messagePtr)

	publicKeyPtr := C.CBytes(publicKey)
	defer C.free(publicKeyPtr)

	//the max length of the message to be encrypted is dictated by the pdo crypto lib (see above)
	if len(message) > int(C.RSA_PLAINTEXT_LEN) {
		return nil, fmt.Errorf("message message too long for encryption")
	}
	//TODO add tests with different message lengths
	encryptedMessageSize := C.RSA_KEY_SIZE >> 3 //bits-to-bytes conversion
	encryptedMessagePtr := C.malloc(C.ulong(encryptedMessageSize))
	defer C.free(encryptedMessagePtr)

	encryptedMessageActualSize := C.uint32_t(0)

	ret := C.pk_encrypt_message(
		(*C.uint8_t)(publicKeyPtr),
		C.uint32_t(len(publicKey)),
		(*C.uint8_t)(messagePtr),
		C.uint32_t(len(message)),
		(*C.uint8_t)(encryptedMessagePtr),
		C.uint32_t(encryptedMessageSize),
		&encryptedMessageActualSize)
	if !ret {
		return nil, fmt.Errorf("encryption failed")
	}

	return C.GoBytes(encryptedMessagePtr, C.int(encryptedMessageActualSize)), nil
}

// DecryptMessage is  symmetric-key encryption performed with the pdo crypto library
func (c PDOCrypto) DecryptMessage(key []byte, encryptedMessage []byte) ([]byte, error) {

	encryptedMessagePtr := C.CBytes(encryptedMessage)
	defer C.free(encryptedMessagePtr)

	keyPtr := C.CBytes(key)
	defer C.free(keyPtr)

	// the (decrypted) message size is estimated to be <= the encrypted message size
	messageSize := len(encryptedMessage)
	messagePtr := C.malloc(C.ulong(messageSize))
	defer C.free(messagePtr)

	messageActualSize := C.uint32_t(0)

	ret := C.decrypt_message(
		(*C.uint8_t)(keyPtr),
		C.uint32_t(len(key)),
		(*C.uint8_t)(encryptedMessagePtr),
		C.uint32_t(len(encryptedMessage)),
		(*C.uint8_t)(messagePtr),
		C.uint32_t(messageSize),
		&messageActualSize)
	if !ret {
		return nil, fmt.Errorf("decryption failed")
	}

	return C.GoBytes(messagePtr, C.int(messageActualSize)), nil
}

//EncryptMessage is a symmetric-key encryption performed with the PDO crypto lib
func (c PDOCrypto) EncryptMessage(key []byte, message []byte) (encryptedMessage []byte, e error) {

	keyPtr := C.CBytes(key)
	defer C.free(keyPtr)

	messagePtr := C.CBytes(message)
	defer C.free(messagePtr)

	//The PDO lib includes the iv and tag in the encrypted message
	encryptedMessageLen := C.uint32_t(len(message)) + C.IV_LEN + C.TAG_LEN
	encryptedMessagePtr := C.malloc(C.ulong(encryptedMessageLen))
	defer C.free(encryptedMessagePtr)

	encryptedMessageActualLen := C.uint32_t(0)

	ret := C.encrypt_message(
		(*C.uint8_t)(keyPtr),
		(C.uint32_t)(len(key)),
		(*C.uint8_t)(messagePtr),
		(C.uint32_t)(len(message)),
		(*C.uint8_t)(encryptedMessagePtr),
		encryptedMessageLen,
		&encryptedMessageActualLen,
	)
	if !ret {
		return nil, fmt.Errorf("encryption failed")
	}

	return C.GoBytes(encryptedMessagePtr, C.int(encryptedMessageActualLen)), nil
}
