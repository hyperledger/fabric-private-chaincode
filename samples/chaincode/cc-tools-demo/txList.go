package main

import (
	txdefs "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/txdefs"

	tx "github.com/hyperledger-labs/cc-tools/transactions"
)

var txList = []tx.Transaction{
	tx.CreateAsset,
	tx.UpdateAsset,
	tx.DeleteAsset,
	txdefs.CreateNewLibrary,
	txdefs.GetNumberOfBooksFromLibrary,
	txdefs.UpdateBookTenant,
	txdefs.GetBooksByAuthor,
}
