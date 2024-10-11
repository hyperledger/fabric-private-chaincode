/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-private-chaincode/samples/application/ccapi/pkg"
)

func InitEnclave() error {
	admin := pkg.NewAdmin(config)
	defer admin.Close()
	return admin.InitEnclave(config.CorePeerId)
}
