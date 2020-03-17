/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("ercc")

func main() {
	// start chaincode
	// err := shim.Start(NewTestErcc())
	err := shim.Start(ercc.NewErcc())
	if err != nil {
		logger.Errorf("Error starting registry chaincode: %s", err)
	}
}
