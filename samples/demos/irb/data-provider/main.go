/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"

	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
)

func testUpload() bool {
	ek, dk, err := crypto.NewRSAKeys()
	if err != nil {
		fmt.Printf("no new RSA")
		return false
	}
	ciao := "ciao"

	m, err := crypto.PkEncryptMessage(ek, []byte(ciao))
	if err != nil {
		fmt.Printf("no encr")
		return false
	}

	s, err := crypto.PkDecryptMessage(dk, m)
	if err != nil {
		fmt.Printf("no decr")
		return false
	}
	fmt.Printf("%s\n", string(s))

	_, err = Upload([]byte(ciao))
	if err != nil {
		fmt.Printf("no upload")
		return false
	}
	fmt.Printf("upload success\n")

	return true
}

func main() {
	//IMPLEMENT ME
}
