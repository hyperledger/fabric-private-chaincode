/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"net/http"

	"github.com/hyperledger/fabric-private-chaincode/samples/application/ccapi/common"
)

func QueryFpcDefault(txName string, args [][]byte) ([]byte, int, error) {
	stringArgs := make([]string, len(args))
	for i, b := range args {
		stringArgs[i] = string(b)
	}

	client, err := common.NewDefaultFpcClient()
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	defer client.Close()

	res, err := client.Query(txName, stringArgs...)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return []byte(res), http.StatusOK, nil
}
