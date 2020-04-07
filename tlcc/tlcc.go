/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tlcc

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger-labs/fabric-private-chaincode/tlcc/enclave"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/common/ledger"
	"github.com/hyperledger/fabric/core/config"
	"github.com/hyperledger/fabric/core/peer"
)

var logger = flogging.MustGetLogger("tlcc")

type TrustedLedgerCC struct {
	enclave enclave.Stub
	peer    *peer.Peer
}

func New(p *peer.Peer) *TrustedLedgerCC {
	return &TrustedLedgerCC{
		enclave: &enclave.StubImpl{},
		peer:    p,
	}
}

func (t *TrustedLedgerCC) Name() string              { return "tlcc" }
func (t *TrustedLedgerCC) Chaincode() shim.Chaincode { return t }

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
		err_msg := fmt.Sprintf("t.enclave.GetLocalAttestationReport failed: %s", err)
		logger.Errorf(err_msg)
		return shim.Error(err_msg)
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
		err_msg := fmt.Sprintf("t.enclave.GetStateMetadata failed: %s", err)
		logger.Errorf(err_msg)
		return shim.Error(err_msg)
	}

	cmacBase64 := base64.StdEncoding.EncodeToString(cmac)
	return shim.Success([]byte(cmacBase64))
}

func (t *TrustedLedgerCC) joinChannel(stub shim.ChaincodeStubInterface) pb.Response {
	channelName := stub.GetChannelID()

	peerLedger := t.peer.GetLedger(channelName)
	if peerLedger == nil {
		return shim.Error(fmt.Sprintf("Cannot open %s ledger", channelName))
	}

	iter, err := peerLedger.GetBlocksIterator(0)
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
		err_msg := fmt.Sprintf("t.enclave.Create failed: %s", err)
		logger.Errorf(err_msg)
		return fmt.Errorf(err_msg)
	}

	err = t.enclave.InitWithGenesis([]byte(genesis))
	if err != nil {
		err_msg := fmt.Sprintf("t.enclave.InitWithGenesis failed: %s", err)
		logger.Errorf(err_msg)
		return fmt.Errorf(err_msg)
	}

	return nil
}
