// +build !fpc

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	cc "github.com/hyperledger-labs/fabric-private-chaincode/demo/chaincode/golang"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

func NewMockAuction() shim.Chaincode {
	return cc.NewMockAuction()
}
