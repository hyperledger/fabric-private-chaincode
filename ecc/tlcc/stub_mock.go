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

package tlcc

import (
	"bytes"

	"github.com/hyperledger/fabric/core/chaincode/shim"
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
