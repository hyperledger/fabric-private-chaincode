/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"net/http"

	"github.com/hyperledger-labs/ccapi/common"
)

func QueryFpc(chaincodeName string, channelName string, txName string, args [][]byte) ([]byte, int, error) {
	stringArgs := make([]string, len(args))
	for i, b := range args {
		stringArgs[i] = string(b)
	}

	client := common.NewFpcClient(chaincodeName, channelName)
	res := client.Query(txName, stringArgs[0:]...)
	return []byte(res), http.StatusOK, nil
}
