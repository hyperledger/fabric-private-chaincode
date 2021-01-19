/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"fmt"
)

// #cgo CFLAGS: -I${SRCDIR}/../../common/crypto
// #cgo LDFLAGS: -L${SRCDIR}/../../common/crypto/_build -L${SRCDIR}/../../common/logging/_build -Wl,--start-group -lupdo-crypto-adapt -lupdo-crypto -Wl,--end-group -lcrypto -lulogging -lstdc++
// #include <stdio.h>
// #include <stdlib.h>
// #include <stdbool.h>
// #include <stdint.h>
// #include "pdo-crypto-c-wrapper.h"
import "C"

func NewRSAKeys() (publicKey []byte, privateKey []byte, e error) {
	//The RSA keys created by the PDO crypto lib are 2048bit long and PEM encoded
	//Here we roughly estimate that they fit in 2KB
	const estimatedPemRsaLen = 2048
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
	if ret == false {
		return nil, nil, fmt.Errorf("cannot create RSA keys")
	}

	return C.GoBytes(serializedPublicKeyPtr, C.int(serializedPublicKeyActualLen)), C.GoBytes(serializedPrivateKeyPtr, C.int(serializedPrivateKeyActualLen)), nil
}

func NewECDSAKeys() (publicKey []byte, privateKey []byte, e error) {
	//The ECDSA keys created by the PDO crypto lib are PEM encoded
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
	if ret == false {
		return nil, nil, fmt.Errorf("cannot create ECDSA keys")
	}

	return C.GoBytes(serializedPublicKeyPtr, C.int(serializedPublicKeyActualLen)), C.GoBytes(serializedPrivateKeyPtr, C.int(serializedPrivateKeyActualLen)), nil
}

func SignMessage(privateKey []byte, message []byte) (signature []byte, e error) {
	privateKeyPtr := C.CBytes(privateKey)
	defer C.free(privateKeyPtr)

	messagePtr := C.CBytes(message)
	defer C.free(messagePtr)

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
	if ret == false {
		return nil, fmt.Errorf("cannot sign message")
	}

	return C.GoBytes(signaturePtr, C.int(signatureActualLen)), nil
}

func PkDecryptMessage(privateKey []byte, encryptedMessage []byte) (message []byte, e error) {
	//This is an RSA dencryption performed with the pdo crypto library
	//Importantly, the library uses 2048bit RSA keys & OAEP encoding, so the input size can be at most ~200bytes
	//TODO-1: bump up the key length to 3072 to match NIST strength
	//TODO-2: extend procedure for large input sizes (via hybrid encryption)

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
	if ret == false {
		return nil, fmt.Errorf("pk decryption failed")
	}

	return C.GoBytes(decryptedMessagePtr, C.int(decryptedMessageActualLen)), nil
}

func EncryptMessage(key []byte, message []byte) (encryptedMessage []byte, e error) {
	//This is a symmetric-key encryption performed with the PDO crypto lib

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
	if ret == false {
		return nil, fmt.Errorf("encryption failed")
	}

	return C.GoBytes(encryptedMessagePtr, C.int(encryptedMessageActualLen)), nil
}
