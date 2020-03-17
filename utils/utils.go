/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

const MrEnclaveStateKey = "MRENCLAVE"

// Response contains the response data and signature produced by the enclave
type Response struct {
	ResponseData []byte `json:"ResponseData"`
	Signature    []byte `json:"Signature"`
	PublicKey    []byte `json:"PublicKey"`
}

const SEP = "."

func Read(file string) []byte {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	if data == nil {
		panic(fmt.Errorf("File is empty"))
	}
	return data
}

func IsSGXCompositeKey(comp_str, sep string) bool {
	return strings.HasPrefix(comp_str, sep) && strings.HasSuffix(comp_str, sep)
}

func TransformToCompositeKey(stub shim.ChaincodeStubInterface, comp_key, sep string) string {
	comp := SplitSGXCompositeKey(comp_key, sep)
	indexKey, _ := stub.CreateCompositeKey(comp[0], comp[1:])
	return indexKey
}

func TransformToSGX(comp, sep string) string {
	return strings.Replace(comp, "\x00", sep, -1)
}

func SplitSGXCompositeKey(comp_str, sep string) []string {
	// check it has SEP in front and end
	if !IsSGXCompositeKey(comp_str, sep) {
		panic("comp_key has wrong format")
	}
	comp := strings.Split(comp_str, sep)
	return comp[1 : len(comp)-1]
}
