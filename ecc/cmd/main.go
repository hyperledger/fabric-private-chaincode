/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = flogging.MustGetLogger("ecc")

func main() {
	// create enclave chaincode
	t := ecc.NewEcc()
	defer t.Destroy()

	// start chaincode
	if err := shim.Start(t); err != nil {
		logger.Errorf("Error starting ecc: %s", err)
	}
}
