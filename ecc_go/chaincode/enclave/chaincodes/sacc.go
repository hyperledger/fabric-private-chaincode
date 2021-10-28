//go:build sacc
// +build sacc

package chaincodes

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sacc "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/sacc-go"
)

func New() shim.Chaincode {
	return &sacc.SimpleAsset{}
}
