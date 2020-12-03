/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"
	"os"

	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/registry"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type serverConfig struct {
	CCID    string
	Address string
}

func main() {

	c := &registry.Contract{}
	c.Verifier = attestation.NewVerifier()
	c.IEvaluator = &utils.IdentityEvaluator{}
	c.BeforeTransaction = registry.MyBeforeTransaction

	ercc, err := contractapi.NewChaincode(c)
	if err != nil {
		log.Panicf("error create enclave registry chaincode: %s", err)
	}

	ccid := os.Getenv("CHAINCODE_PKG_ID")
	addr := os.Getenv("CHAINCODE_SERVER_ADDRESS")

	if len(ccid) > 0 && len(addr) > 0 {
		// start chaincode as a service
		config := serverConfig{
			CCID:    ccid,
			Address: addr,
		}

		server := &shim.ChaincodeServer{
			CCID:    config.CCID,
			Address: config.Address,
			CC:      ercc,
			TLSProps: shim.TLSProperties{
				Disabled: true,
			},
		}

		log.Printf("starting enclave registry (%s)\n", config.CCID)

		if err := server.Start(); err != nil {
			log.Panicf("error starting enclave registry chaincode: %s", err)
		}
	} else if len(ccid) == 0 && len(addr) == 0 {
		// start the chaincode in the traditional way

		log.Printf("starting enclave registry\n")
		if err := ercc.Start(); err != nil {
			log.Panicf("Error starting registry chaincode: %v", err)
		}
	} else {
        log.Panicf("invalid input parameters")
    }

}
