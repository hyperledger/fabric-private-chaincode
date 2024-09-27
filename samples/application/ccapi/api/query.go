/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"fmt"

	"github.com/hyperledger/fabric-private-chaincode/samples/application/cc-tools-demo/pkg"
)

func QueryTransaction(args []string) {
	client := pkg.NewClient(config)
	res := client.Query(args[0], args[1:]...)
	fmt.Println("~> " + res)

}
