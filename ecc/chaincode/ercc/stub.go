/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package ercc

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
)

type Stub interface {
	QueryEnclaveCredentials(stub shim.ChaincodeStubInterface, channelId, chaincodeId, enclaveId string) (*protos.Credentials, error)
}

type StubImpl struct {
}

func (ercc *StubImpl) QueryEnclaveCredentials(stub shim.ChaincodeStubInterface, channelId, chaincodeId, enclaveId string) (*protos.Credentials, error) {
	args := [][]byte{[]byte("queryEnclaveCredentials"), []byte(chaincodeId), []byte(enclaveId)}

	// check again chaincode definition and enclave registry
	resp := stub.InvokeChaincode("ercc", args, channelId)
	if resp.Status != shim.OK {
		return nil, fmt.Errorf("error: %s", resp.Message)
	}

	return utils.UnmarshalCredentials(string(resp.Payload))
}
