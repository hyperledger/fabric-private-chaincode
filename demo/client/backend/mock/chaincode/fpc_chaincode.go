/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"github.com/hyperledger-labs/fabric-private-chaincode/ecc"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

func NewMockAuction() shim.Chaincode {
	return ecc.CreateMockedECC()
}
