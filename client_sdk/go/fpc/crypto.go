/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package fpc

func Encrypt(input []byte, encryptionKey []byte) ([]byte, error) {
	return input, nil
}

func KeyGen() ([]byte, error) {
	return []byte("fake key"), nil
}

func Decrypt(encryptedResponse []byte, resultEncryptionKey []byte) ([]byte, error) {
	return encryptedResponse, nil
}
