/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"

	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	dp "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/data-provider"
)

func main() {
	users := []string{"user1", "user2", "user3", "user4", "user5", "user6", "user7", "user8", "user9"}

	fmt.Printf("Creating users and data...")
	for i := 0; i < len(users); i++ {
		_, _, _, err := dp.LoadOrCreateUser(users[i])
		if err != nil {
			panic(err)
		}

		_, err = LoadOrCreateData(users[i])
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("done.\n")

	fmt.Printf("Uploading data...")
	var decryptionKeys [][]byte
	var dataHandlers []string
	for i := 0; i < len(users); i++ {
		data, err := LoadOrCreateData(users[i])
		if err != nil {
			panic(err)
		}

		// create new encryption
		sk, err := crypto.NewSymmetricKey()
		if err != nil {
			panic(err)
		}

		// encrypt data before uploading
		encryptedData, err := crypto.EncryptMessage(sk, data)
		if err != nil {
			panic(err)
		}
		decryptionKeys = append(decryptionKeys, sk)

		// upload data
		handler, err := dp.Upload(encryptedData)
		if err != nil {
			panic(err)
		}
		dataHandlers = append(dataHandlers, handler)
	}
	fmt.Printf("done.\n")

	fmt.Printf("Registering data...")
	for i := 0; i < len(users); i++ {
		uuid, _, vk, err := dp.LoadOrCreateUser(users[i])
		if err != nil {
			panic(err)
		}

		err = dp.RegisterData(uuid, vk, decryptionKeys[i], dataHandlers[i])
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("done.\n")
}
