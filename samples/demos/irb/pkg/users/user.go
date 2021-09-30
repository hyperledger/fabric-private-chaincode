/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package users

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/crypto"
)

var usersDir string = "users"
var uuidFileName string = "uuid.txt"
var publicKeyFileName string = "publickey.txt"
var privateKeyFileName string = "privatekey.txt"

func CreateUser(userName string) error {
	err := os.MkdirAll(filepath.Join(usersDir, userName), 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(usersDir, userName, uuidFileName), []byte(userName), 0755)
	if err != nil {
		return err
	}

	cp := crypto.NewGoCrypto()
	vk, sk, err := cp.NewECDSAKeys()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(usersDir, userName, publicKeyFileName), vk, 0755)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(usersDir, userName, privateKeyFileName), sk, 0755)
	if err != nil {
		return err
	}

	return nil
}

func LoadUser(userName string) (uuid []byte, sk []byte, vk []byte, e error) {
	sk, err := ioutil.ReadFile(filepath.Join(usersDir, userName, privateKeyFileName))
	if err != nil {
		return nil, nil, nil, err
	}

	vk, err = ioutil.ReadFile(filepath.Join(usersDir, userName, publicKeyFileName))
	if err != nil {
		return nil, nil, nil, err
	}

	uuid, err = ioutil.ReadFile(filepath.Join(usersDir, userName, uuidFileName))
	if err != nil {
		return nil, nil, nil, err
	}

	return uuid, sk, vk, nil
}

func LoadOrCreateUser(userName string) (uuid []byte, sk []byte, vk []byte, e error) {
	a, b, c, err := LoadUser(userName)
	if err == nil {
		return a, b, c, nil
	}

	err = CreateUser(userName)
	if err != nil {
		return nil, nil, nil, err
	}

	return LoadUser(userName)
}
