/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	api "github.com/hyperledger/fabric-private-chaincode/samples/application/cc-tools-demo/api"
)

func main() {
	api.InitConfig()
	api.InitEnclave()

	//Invoking transactions
	///createNewLibrary
	args := []string{"createNewLibrary", "{\"name\":\"samuel\"}"}
	api.InvokeTransaction(args)
	///createNewLibrary
	args = []string{"createAsset", "{\"asset\":[{\"@assetType\":\"person\",\"id\":\"51027337023\",\"name\":\"samuel\"}]}"}
	api.InvokeTransaction(args)
	///createNewLibrary
	args = []string{"createAsset", "{\"asset\":[{\"@assetType\":\"book\", \"title\": \"Fairy tail\"  ,\"author\":\"Martin\",\"currentTenant\":{\"@assetType\": \"person\", \"@key\": \"person:f6c10e69-32ae-5dfb-b17e-9eda4a039cee\"}}]}"}
	api.InvokeTransaction(args)

}
