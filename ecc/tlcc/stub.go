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
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = flogging.MustGetLogger("enclave_stub")

var _logger = func(in string) {
	logger.Info(in)
}

// TLCCStub interface
type TLCCStub interface {
	GetReport(stub shim.ChaincodeStubInterface, chaincodeName, channel string, targetInfo []byte) ([]byte, []byte, error)
	VerifyState(stub shim.ChaincodeStubInterface, chaincodeName, channel, key string, nonce []byte, isRangeQuery bool) ([]byte, error)
}

// TLCCStubImpl implements TLCC interface and calls tlcc
type TLCCStubImpl struct {
}

// RegisterEnclave registers enclave at ercc
func (t *TLCCStubImpl) GetReport(stub shim.ChaincodeStubInterface, chaincodeName, channel string, targetInfo []byte) ([]byte, []byte, error) {
	resp := stub.InvokeChaincode(chaincodeName, [][]byte{[]byte("GET_LOCAL_ATT_REPORT"), targetInfo}, channel)
	if resp.Status != shim.OK {
		return nil, nil, errors.New("Setup failed: Con not register enclave at ercc" + string(resp.Message))
	}

	type Response struct {
		Report    string
		EnclavePk string
	}

	var r Response
	if err := json.Unmarshal(resp.Payload, &r); err != nil {
		return nil, nil, err
	}

	reportBytes, err := base64.StdEncoding.DecodeString(r.Report)
	if err != nil {
		return nil, nil, err
	}

	enclavePkBytes, err := base64.StdEncoding.DecodeString(r.EnclavePk)
	if err != nil {
		return nil, nil, err
	}

	return reportBytes, enclavePkBytes, nil
}

func (t *TLCCStubImpl) VerifyState(stub shim.ChaincodeStubInterface, chaincodeName, channel, key string, nonce []byte, isRangeQuery bool) ([]byte, error) {
	// TODO state prefix currently hardcoded
	prefix := "ecc"
	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)

	var k string
	if isRangeQuery {
		k = prefix + key
	} else {
		k = prefix + "." + key
	}

	resp := stub.InvokeChaincode(chaincodeName, [][]byte{[]byte("VERIFY_STATE"), []byte(k), []byte(nonceBase64), []byte(strconv.FormatBool(isRangeQuery))}, channel)
	if resp.Status != shim.OK {
		return nil, errors.New("Error while performing Verify state" + string(resp.Message))
	}

	// logger.Info("tlccStub got CMAC: " + string(resp.Payload))
	cmacBytes, err := base64.StdEncoding.DecodeString(string(resp.Payload))
	if err != nil {
		return nil, err
	}

	return cmacBytes, nil
}
