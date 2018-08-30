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

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/ledger"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/config"
	"github.com/hyperledger/fabric/core/peer"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"

	"gitlab.zurich.ibm.com/sgx-dev/sgx-cc/tlcc/enclave"
)

var logger = shim.NewLogger("tlcc")

type TrustedLedgerCC struct {
	enclave enclave.Stub
}

func New() shim.Chaincode {
	return &TrustedLedgerCC{
		enclave: &enclave.StubImpl{},
	}
}

func (t *TrustedLedgerCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *TrustedLedgerCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, _ := stub.GetFunctionAndParameters()
	logger.Debug("tlcc: invoke is running " + function)

	if function == "GET_LOCAL_ATT_REPORT" {
		return t.getLocalAttestationReport(stub)
	} else if function == "VERIFY_STATE" {
		return t.getStateMetadata(stub)
	} else if function == "JOIN_CHANNEL" {
		return t.joinChannel(stub)
	}

	jsonResp := "{\"Error\":\" Received unknown function invocation: " + function + "\"}"
	return shim.Error(jsonResp)
}

func (t *TrustedLedgerCC) GetTargetInfo() ([]byte, error) {
	return t.enclave.GetTargetInfo()
}

func (t *TrustedLedgerCC) getLocalAttestationReport(stub shim.ChaincodeStubInterface) pb.Response {
	args := stub.GetStringArgs()
	targetInfo := args[1]

	reportAsBytes, enclavePk, err := t.enclave.GetLocalAttestationReport([]byte(targetInfo))
	if err != nil {
		return shim.Error(fmt.Sprintf("local attestation returns error: %s", err))
	}

	enclavePkBase64 := base64.StdEncoding.EncodeToString(enclavePk)
	reportBase64 := base64.StdEncoding.EncodeToString(reportAsBytes)

	jsonResp := "{\"Report\":\"" + reportBase64 + "\", \"EnclavePK\": \"" + enclavePkBase64 + "\"}"
	return shim.Success([]byte(jsonResp))
}

func (t *TrustedLedgerCC) getStateMetadata(stub shim.ChaincodeStubInterface) pb.Response {
	args := stub.GetStringArgs()
	key := args[1]
	nonce, err := base64.StdEncoding.DecodeString(args[2])
	if err != nil {
		return shim.Error(fmt.Sprintf("Can not parse nonce %s", err))
	}
	isRangeQuery, err := strconv.ParseBool(string(args[3]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Can not parse range query flag %s", err))
	}

	cmac, err := t.enclave.GetStateMetadata(key, []byte(nonce), isRangeQuery)
	if err != nil {
		return shim.Error(fmt.Sprintf("GetState returns error: %s", err))
	}

	cmacBase64 := base64.StdEncoding.EncodeToString(cmac)
	return shim.Success([]byte(cmacBase64))
}

func (t *TrustedLedgerCC) joinChannel(stub shim.ChaincodeStubInterface) pb.Response {
	channelName := stub.GetChannelID()

	ledger := peer.GetLedger(channelName)
	if ledger == nil {
		return shim.Error(fmt.Sprintf("Cannot open %s ledger", channelName))
	}

	iter, err := ledger.GetBlocksIterator(0)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error while getting block iterator, error: %s", err))
	}

	// read first (genesis) block
	res, _ := iter.Next()
	block := res.(*common.Block)
	blockBytes, err := proto.Marshal(block)
	if err != nil {
		panic(err)
	}
	err = t.initNewEnclave(blockBytes)
	if err != nil {
		panic(err)
	}

	// continue reading all blocks in the background
	go t.readBlocks(iter)

	return shim.Success([]byte("Channel joined"))
}

// helper to read all blocks from the ledger and pass them to the enclave
func (t *TrustedLedgerCC) readBlocks(iter ledger.ResultsIterator) {
	for {
		res, _ := iter.Next()
		block := res.(*common.Block)
		blockBytes, err := proto.Marshal(block)
		if err != nil {
			panic(err)
		}

		err = t.enclave.NextBlock(blockBytes)
		if err != nil {
			panic(err)
		}
	}
}

func (t *TrustedLedgerCC) initNewEnclave(genesis []byte) error {
	enclaveLibFile := config.GetPath("sgx.enclave.library")

	// create new Enclave
	err := t.enclave.Create(enclaveLibFile)
	if err != nil {
		return fmt.Errorf("Error while creating enclave %s", err)
	}

	err = t.enclave.InitWithGenesis([]byte(genesis))
	if err != nil {
		return fmt.Errorf("Error while initializing with genesis: %s", err)
	}

	return nil
}

func main() {
	// start chaincode
	err := shim.Start(New())
	if err != nil {
		logger.Errorf("Error starting tlcc: %s", err)
	}
}
