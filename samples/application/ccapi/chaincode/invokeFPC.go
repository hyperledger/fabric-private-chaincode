/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"net/http"

	"github.com/hyperledger-labs/ccapi/common"
)

func InvokeFpc(channelName string, chaincodeName string, txname string, args [][]byte) ([]byte, int, error) {
	stringArgs := make([]string, len(args))
	for i, b := range args {
		stringArgs[i] = string(b)
	}

	client := common.NewFpcClient(channelName, chaincodeName)
	res := client.Invoke(txname, stringArgs[0:]...)
	return []byte(res), http.StatusOK, nil
}
