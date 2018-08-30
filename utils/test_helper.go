/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package utils

import (
	"fmt"
	"plugin"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
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
func CheckInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed:", string(res.Message))
		t.FailNow()
	}
}

func CheckInvoke(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed:", string(res.Message))
		t.FailNow()
	}
}

func CheckState(t *testing.T, stub *shim.MockStub, name string, value string) {
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

func CheckQuery(t *testing.T, stub *shim.MockStub, args [][]byte, value string) {
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

func CheckQueryNotNull(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		t.Fatalf("Query %s failed! Reason: %s", args, string(res.Message))
	}
	if res.Payload == nil {
		t.Fatalf("Query %s failed to get value", args)
	}
}

func CheckStateNotNull(t *testing.T, stub *shim.MockStub, name string) {
	bytes := stub.State[name]
	if bytes == nil {
		t.Fatalf("State %s failed to get value", name)
	}
	if bytes == nil {
		t.Fatalf("State value for  %s is null", name)
	}
}
