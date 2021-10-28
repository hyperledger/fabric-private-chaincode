/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kvtest

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

const MAX_VALUE_SIZE = 1 << 16

type KvTest struct {
}

func (t *KvTest) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *KvTest) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	fmt.Println("KV-Test: +++ Executing chaincode invocation +++")
	functionName, params := stub.GetFunctionAndParameters()
	var result []byte

	if functionName == "put_state" {

		if len(params) != 2 {
			result = []byte("put_state needs 2 parameters: key and value")
		} else {
			if len(params[1]) > MAX_VALUE_SIZE {
				result = []byte("max value size exceeded")
			} else {
				err := stub.PutState(params[0], []byte(params[1]))
				if err != nil {
					return shim.Error(err.Error())
				}
				result = []byte("OK")
			}
		}

	} else if functionName == "get_state" {

		if len(params) != 1 {
			result = []byte("get_state needs 1 parameter: key")
		} else {
			value, err := stub.GetState(params[0])
			if err != nil {
				result = []byte("NOT FOUND")
			}
			result = value
		}

	} else {

		result = []byte("BAD FUNCTION")

	}

	fmt.Println("Response:", string(result))
	fmt.Println("KV-Test: +++ Executing done +++")
	return shim.Success(result)

}
