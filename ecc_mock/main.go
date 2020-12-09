/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"os"

	"github.com/hyperledger-labs/fabric-private-chaincode/ecc_mock/chaincode"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric/common/flogging"
)

type serverConfig struct {
	CCID    string
	Address string
}

var logger = flogging.MustGetLogger("ecc")

func main() {

	// we can control logging via FABRIC_LOGGING_SPEC, the default is FABRIC_LOGGING_SPEC=INFO
	// For more fine grained logging we could also use different log level for loggers.
	// For example: FABRIC_LOGGING_SPEC=ecc=DEBUG:ecc_enclave=ERROR

	// See chaincode.env.example
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_PKG_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	// create enclave chaincode
	ecc := &chaincode.EnclaveChaincode{}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      ecc,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	logger.Infof("starting fpc chaincode (%s)", config.CCID)

	if err := server.Start(); err != nil {
		logger.Panicf("error starting fpc chaincode: %s", err)
	}
}
