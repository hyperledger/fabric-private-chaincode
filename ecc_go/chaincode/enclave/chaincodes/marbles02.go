//go:build marbles02
// +build marbles02

package chaincodes

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	marbles02 "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/marbles02-go"
)

func New() shim.Chaincode {
	return &marbles02.SimpleChaincode{}
}
