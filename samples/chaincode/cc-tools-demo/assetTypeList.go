package main

import (
	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/assettypes"
)

var assetTypeList = []assets.AssetType{
	assettypes.Person,
	assettypes.Book,
	assettypes.Library,
	assettypes.Secret,
}
