/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
	auction "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/auction-go/chaincode"
)

func main() {

	// we can control logging via FABRIC_LOGGING_SPEC, the default is FABRIC_LOGGING_SPEC=INFO
	// For more fine-grained logging we could also use different log level for loggers.
	// For example: FABRIC_LOGGING_SPEC=ecc=DEBUG:ecc_enclave=ERROR

	ccid := os.Getenv("CHAINCODE_PKG_ID")
	addr := os.Getenv("CHAINCODE_SERVER_ADDRESS")

	// create private chaincode
	privateAuction := chaincode.NewPrivateChaincode(&auction.Auction{})

	// start chaincode as a service
	server := &shim.ChaincodeServer{
		CCID:    ccid,
		Address: addr,
		CC:      privateAuction,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	fmt.Printf("starting fpc chaincode (%s)", ccid)
	if err := server.Start(); err != nil {
		panic(err)
	}
}
