/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/enclave"
	"github.com/hyperledger/fabric-private-chaincode/ecc/chaincode/ercc"
	"github.com/hyperledger/fabric-private-chaincode/internal/endorsement"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("ecc")

func main() {

	// we can control logging via FABRIC_LOGGING_SPEC, the default is FABRIC_LOGGING_SPEC=INFO
	// For more fine grained logging we could also use different log level for loggers.
	// For example: FABRIC_LOGGING_SPEC=ecc=DEBUG:ecc_enclave=ERROR

	// create enclave chaincode
	ecc := &chaincode.EnclaveChaincode{
		Enclave:   enclave.NewEnclaveStub(),
		Validator: endorsement.NewValidator(),
		Extractor: &chaincode.ExtractorImpl{},
		Ercc:      &ercc.StubImpl{},
	}

	ccid := os.Getenv("CHAINCODE_PKG_ID")
	addr := os.Getenv("CHAINCODE_SERVER_ADDRESS")

	if len(ccid) > 0 && len(addr) > 0 {
		// start chaincode as a service
		server := &shim.ChaincodeServer{
			CCID:    ccid,
			Address: addr,
			CC:      ecc,
			TLSProps: shim.TLSProperties{
				Disabled: true,
			},
		}

		logger.Infof("starting fpc chaincode (%s)", ccid)

		if err := server.Start(); err != nil {
			logger.Panicf("error starting fpc chaincode: %s", err)
		}
	} else if len(ccid) == 0 && len(addr) == 0 {
		// start the chaincode in the traditional way

		logger.Info("starting enclave registry")
		if err := shim.Start(ecc); err != nil {
			logger.Panicf("Error starting fpc chaincode: %v", err)
		}
	} else {
		logger.Panicf("invalid input parameters")
	}
}
