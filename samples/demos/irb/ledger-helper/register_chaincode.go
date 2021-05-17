/*
   Copyright IBM Corp. All Rights Reserved.
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"path/filepath"

	"github.com/hyperledger/fabric-private-chaincode/integration/client_sdk/go/utils"
	testutils "github.com/hyperledger/fabric-private-chaincode/integration/client_sdk/go/utils"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("register_chaincode")

func main() {
	logger.Infof("Registering chaincode on the ledger...")
	ccID := "experiment-approval-service"
	ccPath := filepath.Join(utils.FPCPath, "samples", "demos", "irb", ccID, "_build", "lib")
	// setup auction chaincode (install, approve, commit)
	initEnclave := true
	err := testutils.Setup(ccID, ccPath, initEnclave)
	if err != nil {
		panic(err)
	}

	logger.Infof("done!")
}
