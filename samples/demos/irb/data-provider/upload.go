/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package data_provider

import (
	"crypto/sha256"
	"encoding/base64"

	storage "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/storage/go"
)

func Upload(data []byte) (handler string, e error) {
	hashedContent := sha256.Sum256(data)
	encodedContent := base64.StdEncoding.EncodeToString(data)
	key := base64.StdEncoding.EncodeToString(hashedContent[:])

	err := storage.Set(key, encodedContent)
	if err != nil {
		return "", err
	}
	return key, nil
}
