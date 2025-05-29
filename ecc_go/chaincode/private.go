/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/ercc"
	"github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode/enclave_go"
	"github.com/hyperledger/fabric-private-chaincode/internal/endorsement"
)

type BuildOption func(*chaincode.EnclaveChaincode, shim.Chaincode)

// NewPrivateChaincode creates a new chaincode! This is for go support only!!!
func NewPrivateChaincode(cc shim.Chaincode, options ...BuildOption) *chaincode.EnclaveChaincode {
	ecc := &chaincode.EnclaveChaincode{
		Enclave:   enclave_go.NewEnclaveStub(cc),
		Validator: endorsement.NewValidator(),
		Extractor: &chaincode.ExtractorImpl{},
		Ercc:      &ercc.StubImpl{},
	}
	for _, o := range options {
		o(ecc, cc)
	}
	return ecc
}

func WithSKVS() BuildOption {
	return func(ecc *chaincode.EnclaveChaincode, cc shim.Chaincode) {
		ecc.Enclave = enclave_go.NewSkvsStub(cc)
	}
}
