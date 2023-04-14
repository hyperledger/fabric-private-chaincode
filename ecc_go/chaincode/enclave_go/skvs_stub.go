/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"crypto/sha256"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type skvsStub struct {
	*EnclaveStub
}

func NewSkvsStub(cc shim.Chaincode) *skvsStub {
	enclaveStub := NewEnclaveStub(cc)
	return &skvsStub{enclaveStub}
}

func (e *skvsStub) ChaincodeInvoke(stub shim.ChaincodeStubInterface, chaincodeRequestMessageBytes []byte) ([]byte, error) {
	logger.Error("==== SKVS ChaincodeInvoke ====")

	signedProposal, err := stub.GetSignedProposal()
	if err != nil {
		shim.Error(err.Error())
	}

	if err := e.verifySignedProposal(stub, chaincodeRequestMessageBytes); err != nil {
		return nil, errors.Wrap(err, "signed proposal verification failed")
	}

	// unmarshal chaincodeRequest
	chaincodeRequestMessage := &protos.ChaincodeRequestMessage{}
	err = proto.Unmarshal(chaincodeRequestMessageBytes, chaincodeRequestMessage)
	if err != nil {
		return nil, err
	}

	// get key transport message including the encryption keys for request and response
	keyTransportMessage, err := e.extractKeyTransportMessage(chaincodeRequestMessage)
	if err != nil {
		return nil, errors.Wrap(err, "cannot extract keyTransportMessage")
	}

	// decrypt request
	cleartextChaincodeRequest, err := e.extractCleartextChaincodeRequest(chaincodeRequestMessage, keyTransportMessage)
	if err != nil {
		return nil, errors.Wrap(err, "cannot decrypt chaincode request")
	}

	// create a new instance of a FPC RWSet that we pass to the stub and later return with the response
	rwset := NewReadWriteSet()

	// Invoke chaincode
	// we wrap the stub with our FpcStubInterface
	// ** Implement our own FpcStubInterface
	skvsStub := NewSkvsStubInterface(stub, cleartextChaincodeRequest.GetInput(), rwset, e.ccKeys)
	ccResponse := e.ccRef.Invoke(skvsStub)
	// **
	// fpcStub := NewFpcStubInterface(stub, cleartextChaincodeRequest.GetInput(), rwset, e.ccKeys)
	// ccResponse := e.ccRef.Invoke(fpcStub)

	// marshal chaincode response
	ccResponseBytes, err := protoutil.Marshal(&ccResponse)
	if err != nil {
		return nil, err
	}

	//encrypt response
	encryptedResponse, err := e.csp.EncryptMessage(keyTransportMessage.GetResponseEncryptionKey(), ccResponseBytes)
	if err != nil {
		return nil, err
	}

	chaincodeRequestMessageHash := sha256.Sum256(chaincodeRequestMessageBytes)

	response := &protos.ChaincodeResponseMessage{
		EncryptedResponse:           encryptedResponse,
		FpcRwSet:                    rwset.ToFPCKVSet(),
		EnclaveId:                   e.identity.GetEnclaveId(),
		Proposal:                    signedProposal,
		ChaincodeRequestMessageHash: chaincodeRequestMessageHash[:],
	}

	responseBytes, err := proto.Marshal(response)
	if err != nil {
		return nil, err
	}

	// create signature
	sig, err := e.identity.Sign(responseBytes)
	if err != nil {
		return nil, err
	}

	signedResponse := &protos.SignedChaincodeResponseMessage{
		ChaincodeResponseMessage: responseBytes,
		Signature:                sig,
	}

	return proto.Marshal(signedResponse)
}
