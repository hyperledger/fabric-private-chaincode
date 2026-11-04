//go:build fpc_cgo

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode"
	"github.com/hyperledger/fabric-private-chaincode/sgx/ecc/enclave"
)

// newEnclave returns the SGX enclave (cgo), built with fpc_cgo.
func newEnclave() chaincode.Enclave { return enclave.NewEnclaveStub() }
