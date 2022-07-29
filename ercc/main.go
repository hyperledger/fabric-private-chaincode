/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-private-chaincode/ercc/attestation"
	"github.com/hyperledger/fabric-private-chaincode/ercc/registry"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("ercc")

func main() {

	// we can control logging via FABRIC_LOGGING_SPEC, the default is FABRIC_LOGGING_SPEC=INFO
	// For more fine grained logging we could also use different log level for loggers.
	// For example: FABRIC_LOGGING_SPEC=ecc=DEBUG:ecc_enclave=ERROR

	c := &registry.Contract{}
	c.Verifier = attestation.GetAvailableVerifier()
	c.IEvaluator = &utils.IdentityEvaluator{}
	c.BeforeTransaction = registry.MyBeforeTransaction

	ercc, err := contractapi.NewChaincode(c)
	if err != nil {
		logger.Panicf("error create enclave registry chaincode: %s", err)
	}

	ccid := os.Getenv("CHAINCODE_PKG_ID")
	addr := os.Getenv("CHAINCODE_SERVER_ADDRESS")

	if len(ccid) > 0 && len(addr) > 0 {
		// start chaincode as a service
		server := &shim.ChaincodeServer{
			CCID:    ccid,
			Address: addr,
			CC:      ercc,
			TLSProps: shim.TLSProperties{
				Disabled: true,
			},
		}

		logger.Infof("starting enclave registry (%s)", ccid)

		if err := server.Start(); err != nil {
			logger.Panicf("error starting enclave registry chaincode: %s", err)
		}
	} else if len(ccid) == 0 && len(addr) == 0 {
		// start the chaincode in the traditional way

		logger.Info("starting enclave registry")
		if err := ercc.Start(); err != nil {
			logger.Panicf("Error starting registry chaincode: %v", err)
		}
	} else {
		logger.Panicf("invalid input parameters")
	}

}
