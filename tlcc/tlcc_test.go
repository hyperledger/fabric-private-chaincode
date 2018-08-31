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

package main

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/peer"
	"github.com/spf13/viper"
	"github.com/hyperledger-labs/fabric-secure-chaincode/tlcc/enclave"
	th "github.com/hyperledger-labs/fabric-secure-chaincode/utils"
)

func setupTestLedger(chainid string) {
	viper.Set("peer.fileSystemPath", "/tmp/hyperledger/test/")
	peer.MockInitialize()
	peer.MockCreateChain(chainid)
}

func TestTrustedLedgerCC_Init(t *testing.T) {
	tlcc := createTlcc()
	stub := shim.NewMockStub("tlcc", tlcc)
	setupTestLedger("mychannel")
	th.CheckInit(t, stub, [][]byte{})
}

func TestTrustedLedgerCC_JoinChannel(t *testing.T) {
	tlcc := createTlcc()
	stub := shim.NewMockStub("tlcc", tlcc)

	setupTestLedger("mychannel")
	th.CheckInit(t, stub, [][]byte{})
	th.CheckInvoke(t, stub, [][]byte{[]byte("JOIN_CHANNEL"), []byte("mychannel")})
}

func TestTrustedLedgerCC_GetReport(t *testing.T) {
	tlcc := createTlcc()
	stub := shim.NewMockStub("tlcc", tlcc)

	setupTestLedger("mychannel")
	th.CheckInit(t, stub, [][]byte{})
	th.CheckInvoke(t, stub, [][]byte{[]byte("JOIN_CHANNEL"), []byte("mychannel")})

	// note this is a debugging call :D
	targetInfo, _ := tlcc.GetTargetInfo()

	args := [][]byte{[]byte("GET_LOCAL_ATT_REPORT"), targetInfo}
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.FailNow()
	}
}

func TestTrustedLedgerCC_GetStateCMAC(t *testing.T) {
	tlcc := createTlcc()
	stub := shim.NewMockStub("tlcc", tlcc)

	setupTestLedger("mychannel")
	th.CheckInit(t, stub, [][]byte{})
	th.CheckInvoke(t, stub, [][]byte{[]byte("JOIN_CHANNEL"), []byte("mychannel")})

	key := []byte("some.channel.someKey")
	nonce := []byte(base64.StdEncoding.EncodeToString([]byte("moin")))
	isRangeQuery := []byte(strconv.FormatBool(false))

	args := [][]byte{[]byte("VERIFY_STATE"), key, nonce, isRangeQuery}

	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.FailNow()
	}
	fmt.Println("CMAC: " + string(res.Payload))
}

func TestLoadPlugin(t *testing.T) {
	th.CheckLoadPlugin(t, "tlcc.so")
}

func createTlcc() *TrustedLedgerCC {
	return &TrustedLedgerCC{
		enclave: &enclave.MockStub{},
	}
}
