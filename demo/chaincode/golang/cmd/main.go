/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	cc "github.com/hyperledger-labs/fabric-private-chaincode/demo/chaincode/golang"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = flogging.MustGetLogger("go_mock_auction")

func main() {
	// create enclave chaincode
	t := cc.NewMockAuction()

	// start chaincode
	if err := shim.Start(t); err != nil {
		logger.Errorf("Error starting ecc: %s", err)
	}
}
