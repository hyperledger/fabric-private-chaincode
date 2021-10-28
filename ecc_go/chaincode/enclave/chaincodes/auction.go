//go:build auction
// +build auction

package chaincodes

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	auction "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/auction-go"
)

func New() shim.Chaincode {
	return &auction.Auction{}
}
