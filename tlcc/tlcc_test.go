/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tlcc

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/hyperledger-labs/fabric-private-chaincode/tlcc/enclave"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric/core/ledger/ledgermgmt"
	"github.com/hyperledger/fabric/core/ledger/ledgermgmt/ledgermgmttest"
	"github.com/hyperledger/fabric/core/peer"
)

const ChannelId = "mychannel"

func setupTestLedger(cid string) (*shimtest.MockStub, *TrustedLedgerCC, func()) {

	peerInstance := &peer.Peer{}

	tempdir, err := ioutil.TempDir("", "tlcc-test")
	if err != nil {
		panic(fmt.Sprintf("failed to create temporary directory: %s", err))
	}

	initializer := ledgermgmttest.NewInitializer(filepath.Join(tempdir, "ledgerData"))
	peerInstance.LedgerMgr = ledgermgmt.NewLedgerMgr(initializer)

	cleanup := func() {
		peerInstance.LedgerMgr.Close()
		os.RemoveAll(tempdir)
	}

	err = peer.CreateMockChannel(peerInstance, cid, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create channel: %s", err))
	}

	tlcc := &TrustedLedgerCC{
		enclave: &enclave.MockStub{},
		peer:    peerInstance,
	}

	stub := shimtest.NewMockStub("tlcc", tlcc)
	stub.ChannelID = cid
	return stub, tlcc, cleanup
}

func TestTrustedLedgerCC_Init(t *testing.T) {
	stub, _, cleanup := setupTestLedger(ChannelId)
	defer cleanup()
	CheckInit(t, stub, [][]byte{})
}

func TestTrustedLedgerCC_JoinChannel(t *testing.T) {
	stub, _, cleanup := setupTestLedger(ChannelId)
	defer cleanup()
	CheckInit(t, stub, [][]byte{})
	CheckInvoke(t, stub, [][]byte{[]byte("JOIN_CHANNEL"), []byte(ChannelId)})
}

func TestTrustedLedgerCC_GetReport(t *testing.T) {
	stub, tlcc, cleanup := setupTestLedger(ChannelId)
	defer cleanup()
	CheckInit(t, stub, [][]byte{})
	CheckInvoke(t, stub, [][]byte{[]byte("JOIN_CHANNEL"), []byte(ChannelId)})

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
	stub, _, cleanup := setupTestLedger(ChannelId)
	defer cleanup()
	CheckInit(t, stub, [][]byte{})
	CheckInvoke(t, stub, [][]byte{[]byte("JOIN_CHANNEL"), []byte(ChannelId)})

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
