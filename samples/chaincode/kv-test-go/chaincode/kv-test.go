/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

const MAX_VALUE_SIZE = 1 << 16

type KvTest struct {
	functionRegister map[string]func(stubInterface shim.ChaincodeStubInterface) pb.Response
}

func NewKvTest() *KvTest {
	cc := &KvTest{
		functionRegister: make(map[string]func(stubInterface shim.ChaincodeStubInterface) pb.Response),
	}

	cc.functionRegister["put_state"] = putState
	cc.functionRegister["get_state"] = getState
	cc.functionRegister["del_state"] = delState

	return cc
}

func (t *KvTest) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *KvTest) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("KV-Test: +++ Executing chaincode invocation +++")
	defer fmt.Println("KV-Test: +++ Executing done +++")

	return t.dispatch(stub)
}

func (t *KvTest) dispatch(stub shim.ChaincodeStubInterface) pb.Response {
	functionType, params := stub.GetFunctionAndParameters()
	if f, exist := t.functionRegister[functionType]; exist {
		fmt.Printf("call f='%s' with args='%v'\n", functionType, params)
		return f(stub)
	}

	return shim.Error(fmt.Sprintf("function '%s' not known", functionType))
}

func getState(stub shim.ChaincodeStubInterface) pb.Response {
	_, params := stub.GetFunctionAndParameters()

	if len(params) != 1 {
		return shim.Success([]byte("get_state needs 1 parameter: key"))
	}

	value, err := stub.GetState(params[0])
	if err != nil {
		return shim.Success([]byte("not found"))
	}

	return shim.Success(value)
}

func putState(stub shim.ChaincodeStubInterface) pb.Response {
	_, params := stub.GetFunctionAndParameters()

	if len(params) != 2 {
		return shim.Success([]byte("put_state needs 2 parameters: key and value"))
	}

	if len(params[1]) > MAX_VALUE_SIZE {
		return shim.Success([]byte("max value size exceeded"))
	}

	err := stub.PutState(params[0], []byte(params[1]))
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}

	return shim.Success([]byte("ok"))
}

func delState(stub shim.ChaincodeStubInterface) pb.Response {
	_, params := stub.GetFunctionAndParameters()

	if len(params) != 1 {
		return shim.Success([]byte("del_state needs 1 parameter: key"))
	}

	err := stub.DelState(params[0])
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}

	return shim.Success([]byte("ok"))
}
