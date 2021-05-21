/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package data_provider

import (
	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	storage "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/storage/go"
)

func Upload(data []byte) (privateKey []byte, e error) {
	ek, dk, err := crypto.NewRSAKeys()
	if err != nil {
		return nil, err
	}

	encryptedData, err := crypto.PkEncryptMessage(ek, []byte(data))
	if err != nil {
		return nil, err
	}

	err = storage.Set(ek, encryptedData)
	if err != nil {
		return nil, err
	}

	return dk, nil
}
