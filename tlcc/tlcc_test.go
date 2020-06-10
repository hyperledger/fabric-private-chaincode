/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tlcc

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"

	. "github.com/hyperledger-labs/fabric-private-chaincode/internal/testing"
	"github.com/hyperledger-labs/fabric-private-chaincode/tlcc/enclave"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/spf13/viper"
)

func setupTestLedger(chainid string) {
	viper.Set("peer.fileSystemPath", "/tmp/hyperledger/test/")
	// TODO
	// MockInitialize and MockCreateChain have been removed with Fabric v2. To make golinter happy again this is
	// simply commented out as this test is not invoked by make script anyway. Another PR will re-enable this test.
	// peer.MockInitialize()
	// peer.MockCreateChain(chainid)
}

func TestTrustedLedgerCC_Init(t *testing.T) {
	tlcc := createTlcc()
	stub := shimtest.NewMockStub("tlcc", tlcc)
	stub.ChannelID = "mychannel"
	setupTestLedger("mychannel")
	CheckInit(t, stub, [][]byte{})
}

func TestTrustedLedgerCC_JoinChannel(t *testing.T) {
	tlcc := createTlcc()
	stub := shimtest.NewMockStub("tlcc", tlcc)
	stub.ChannelID = "mychannel"

	setupTestLedger("mychannel")
	CheckInit(t, stub, [][]byte{})
	CheckInvoke(t, stub, [][]byte{[]byte("JOIN_CHANNEL"), []byte("mychannel")})
}

func TestTrustedLedgerCC_GetReport(t *testing.T) {
	tlcc := createTlcc()
	stub := shimtest.NewMockStub("tlcc", tlcc)
	stub.ChannelID = "mychannel"

	setupTestLedger("mychannel")
	CheckInit(t, stub, [][]byte{})
	CheckInvoke(t, stub, [][]byte{[]byte("JOIN_CHANNEL"), []byte("mychannel")})

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
	stub := shimtest.NewMockStub("tlcc", tlcc)
	stub.ChannelID = "mychannel"

	setupTestLedger("mychannel")
	CheckInit(t, stub, [][]byte{})
	CheckInvoke(t, stub, [][]byte{[]byte("JOIN_CHANNEL"), []byte("mychannel")})

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
	CheckLoadPlugin(t, "tlcc.so")
}

func createTlcc() *TrustedLedgerCC {
	return &TrustedLedgerCC{
		enclave: &enclave.MockStub{},
	}
}
