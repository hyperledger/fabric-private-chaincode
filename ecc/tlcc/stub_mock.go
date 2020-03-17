/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tlcc

import (
	"bytes"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

var spid = [16]byte{0x25, 0x42, 0xF9, 0x0D, 0x63, 0x31, 0x2C, 0x1D, 0xA2, 0xF6, 0xE2, 0x18, 0x76, 0xF0, 0xE3, 0x89}

// MockTLCCStubImpl implements TLCC interface and calls tlcc
type MockTLCCStub struct {
}

func (t *MockTLCCStub) GetReport(stub shim.ChaincodeStubInterface, chaincodeName, channel string, targetInfo []byte) ([]byte, []byte, error) {
	return nil, nil, nil
}

func (t *MockTLCCStub) VerifyState(stub shim.ChaincodeStubInterface, chaincodeName, channel, key string, nonce []byte, isRangeQuer bool) ([]byte, error) {
	cmac := make([]byte, 16)
	cmac = bytes.Repeat([]byte{0xff}, 16)
	return cmac, nil
}
