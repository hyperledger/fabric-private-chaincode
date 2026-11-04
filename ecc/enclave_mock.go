//go:build !fpc_cgo

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/enclave"
)

// newEnclave returns the mock enclave: without fpc_cgo there is no SGX enclave.
func newEnclave() chaincode.Enclave { return enclave.NewEnclaveStub() }
