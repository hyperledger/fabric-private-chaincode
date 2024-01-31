/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
test
test2
test3*/

package chaincode

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type Echo struct {
}

func (t *Echo) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *Echo) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	functionName, params := stub.GetFunctionAndParameters()
	fmt.Println("EchoCC: Function:", functionName, "Params:", params)
	return shim.Success([]byte(functionName))
}
