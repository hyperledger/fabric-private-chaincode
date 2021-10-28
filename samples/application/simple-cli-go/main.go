/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger/fabric-private-chaincode/samples/application/simple-cli-go/cmd"
	"time"
	"fmt"
)

func main() {
	start := time.Now()
	cmd.Execute()
	fmt.Println(">", time.Since(start))
}
