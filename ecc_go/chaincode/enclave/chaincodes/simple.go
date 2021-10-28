//go:build simple
// +build simple

package chaincodes

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/simple-go"
)

func New() shim.Chaincode {
	return &simple.SimpleChaincode{}
}
