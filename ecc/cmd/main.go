/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("ecc")

func main() {

	// we can control logging via FABRIC_LOGGING_SPEC, the default is FABRIC_LOGGING_SPEC=INFO
	// For more fine grained logging we could also use different log level for loggers.
	// For example: FABRIC_LOGGING_SPEC=ecc=DEBUG:ecc_enclave=ERROR

	// create enclave chaincode
	t := ecc.NewEcc()
	defer t.Destroy()

	// start chaincode
	if err := shim.Start(t); err != nil {
		logger.Errorf("Error starting ecc: %s", err)
	}
}
