/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"os"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-lib-go/common/flogging"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/ercc"
	"github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode/enclave_go"
	"github.com/hyperledger/fabric-private-chaincode/internal/endorsement"
)

var logger = flogging.MustGetLogger("ecc_go")

// noSGX is set to "true" at link time by builds that don't run in an enclave
// (`make with_go`, -ldflags "-X .../ecc_go/chaincode.noSGX=true").
var noSGX string

// warnIfNoSGX logs a warning when the chaincode runs without SGX protection:
// a plain go build (noSGX) or ego in SGX simulation mode (OE_SIMULATION=1).
func warnIfNoSGX(noSGX, oeSimulation string, log func(string)) {
	if noSGX == "true" || oeSimulation == "1" {
		log("FPC running without SGX: no confidentiality, for development only")
	}
}

type BuildOption func(*chaincode.EnclaveChaincode, shim.Chaincode)

// NewPrivateChaincode creates a new chaincode! This is for go support only!!!
func NewPrivateChaincode(cc shim.Chaincode, options ...BuildOption) *chaincode.EnclaveChaincode {
	warnIfNoSGX(noSGX, os.Getenv("OE_SIMULATION"), func(msg string) { logger.Warning(msg) })

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
