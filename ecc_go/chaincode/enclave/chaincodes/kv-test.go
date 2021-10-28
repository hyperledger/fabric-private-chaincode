//go:build kvtest
// +build kvtest

package chaincodes

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	kvtest "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/kv-test-go"
)

func New() shim.Chaincode {
	return &kvtest.KvTest{}
}
