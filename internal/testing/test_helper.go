/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package testing

import (
	"fmt"
	"plugin"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
)

func CheckLoadPlugin(t *testing.T, path string) {
	p, err := plugin.Open(path)
	if err != nil {
		t.Error(err)
	}

	factorySymbol, err := p.Lookup("New")
	if err != nil {
		t.Error(err)
	}

	factory, ok := factorySymbol.(func() shim.Chaincode)
	if !ok {
		t.FailNow()
	}

	factory()
}

// helper methods
func CheckInit(t *testing.T, stub *shimtest.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed:", string(res.Message))
		t.FailNow()
	}
}

func CheckInvoke(t *testing.T, stub *shimtest.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed:", string(res.Message))
		t.FailNow()
	}
}

func CheckState(t *testing.T, stub *shimtest.MockStub, name string, value string) {
	bytes := stub.State[name]
	if bytes == nil {
		fmt.Println("State", name, "failed to get value")
		t.FailNow()
	}
	if string(bytes) != value {
		fmt.Println("State value", name, "was not", value, "as expected")
		t.FailNow()
	}
}

func CheckQuery(t *testing.T, stub *shimtest.MockStub, args [][]byte, value string) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		t.Fatalf("Query %s failed! Reason: %s", args, string(res.Message))
	}
	if res.Payload == nil {
		t.Fatalf("Query %s failed to get value", args)
	}
	if string(res.Payload) != value {
		t.Fatalf("Query value %s was not %s as expected", args, value)
	}
}

func CheckQueryNotNull(t *testing.T, stub *shimtest.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		t.Fatalf("Query %s failed! Reason: %s", args, string(res.Message))
	}
	if res.Payload == nil {
		t.Fatalf("Query %s failed to get value", args)
	}
}

func CheckStateNotNull(t *testing.T, stub *shimtest.MockStub, name string) {
	bytes := stub.State[name]
	if bytes == nil {
		t.Fatalf("State %s failed to get value", name)
	}
	if bytes == nil {
		t.Fatalf("State value for  %s is null", name)
	}
}
