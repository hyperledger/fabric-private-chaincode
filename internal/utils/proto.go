/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	//lint:ignore SA1019 old protos are needed for fabric
	protoV1 "github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// MarshallProtoBase64 returns a serialized protobuf message encoded as base64 string
func MarshallProtoBase64(msg proto.Message) string {
	return base64.StdEncoding.EncodeToString(MarshalOrPanic(msg))
}

// MarshallProto returns a serialized protobuf message
func MarshallProto(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

func MarshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

func UnmarshalCredentials(credentialsBase64 string) (*protos.Credentials, error) {
	credentialsBytes, err := base64.StdEncoding.DecodeString(credentialsBase64)
	if err != nil {
		return nil, err
	}

	if len(credentialsBytes) == 0 {
		return nil, fmt.Errorf("credential input empty")
	}

	credentials := &protos.Credentials{}
	err = proto.Unmarshal(credentialsBytes, credentials)
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func UnmarshalAttestedData(serializedAttestedData *anypb.Any) (*protos.AttestedData, error) {
	if serializedAttestedData == nil {
		return nil, errors.New("attested data is empty")
	}

	attestedData := &protos.AttestedData{}
	if err := serializedAttestedData.UnmarshalTo(attestedData); err != nil {
		return nil, errors.Wrap(err, "invalid attested data message")
	}

	return attestedData, nil
}

func UnmarshalInitEnclaveMessage(data []byte) (*protos.InitEnclaveMessage, error) {
	if data == nil {
		return nil, errors.New("initEnclaveMessage is empty")
	}

	msg := &protos.InitEnclaveMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, errors.Wrap(err, "invalid attested data message")
	}

	return msg, nil
}

func UnmarshalQueryChaincodeDefinitionResult(data []byte) (*lifecycle.QueryChaincodeDefinitionResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("QueryChaincodeDefinitionResult is empty")
	}

	df := &lifecycle.QueryChaincodeDefinitionResult{}
	if err := proto.Unmarshal(data, protoV1.MessageV2(df)); err != nil {
		return nil, errors.Wrap(err, "invalid QueryChaincodeDefinitionResult")
	}
	return df, nil
}

func UnmarshalSignedChaincodeResponseMessage(data []byte) (*protos.SignedChaincodeResponseMessage, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("SignedChaincodeResponseMessage is empty")
	}

	msg := &protos.SignedChaincodeResponseMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, errors.Wrap(err, "invalid SignedChaincodeResponseMessage")
	}

	return msg, nil
}

func UnmarshalChaincodeResponseMessage(data []byte) (*protos.ChaincodeResponseMessage, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("ChaincodeResponseMessage data empty")
	}

	msg := &protos.ChaincodeResponseMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, errors.Wrap(err, "invalid ChaincodeResponseMessage")
	}

	return msg, nil
}

// GetEnclaveId returns enclave_id as hex-encoded string of SHA256 hash over enclave_vk.
func GetEnclaveId(attestedData *protos.AttestedData) string {
	// hash enclave vk
	h := sha256.Sum256(attestedData.EnclaveVk)
	// encode and normalize
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

func ExtractEndpoint(credentials *protos.Credentials) (string, error) {
	attestedData := &protos.AttestedData{}
	err := credentials.SerializedAttestedData.UnmarshalTo(attestedData)
	if err != nil {
		return "", err
	}

	return attestedData.HostParams.PeerEndpoint, nil
}

func GetChaincodeRequestMessageFromSignedProposal(signedProposal *pb.SignedProposal) (crmProtoBytes []byte, e error) {
	// This function is based on the `newChaincodeStub` in
	// https://github.com/hyperledger/fabric-chaincode-go/blob/9b3ae92d8664f398fe9846561fafde44864d55e3/shim/stub.go
	// In particular, it serves to
	// * extract the arguments from the proposal,
	// * check that exactly two arguments have been passed (i.e., the function name and crm)
	// * return the (base-64 decoded) chaincode request message bytes

	if signedProposal == nil {
		return nil, fmt.Errorf("no signed proposal to parse")
	}

	var err error

	proposal, err := protoutil.UnmarshalProposal(signedProposal.ProposalBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to extract Proposal from SignedProposal: %s", err)
	}

	// check for header
	if len(proposal.GetHeader()) == 0 {
		return nil, errors.New("failed to extract Proposal fields: proposal header is nil")
	}

	// extract header
	hdr, err := protoutil.UnmarshalHeader(proposal.GetHeader())
	if err != nil {
		return nil, fmt.Errorf("failed to extract proposal header: %s", err)
	}

	// validate channel header
	chdr, err := protoutil.UnmarshalChannelHeader(hdr.ChannelHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract channel header: %s", err)
	}
	validTypes := map[common.HeaderType]bool{
		common.HeaderType_ENDORSER_TRANSACTION: true,
	}
	if !validTypes[common.HeaderType(chdr.GetType())] {
		return nil, fmt.Errorf(
			"invalid channel header type. Expected %s, received %s",
			common.HeaderType_ENDORSER_TRANSACTION,
			common.HeaderType(chdr.GetType()),
		)
	}

	// extract args from proposal payload
	payload, err := protoutil.UnmarshalChaincodeProposalPayload(proposal.GetPayload())
	if err != nil {
		return nil, fmt.Errorf("failed to extract proposal payload: %s", err)
	}
	cppInput := payload.GetInput()
	if cppInput == nil {
		return nil, fmt.Errorf("failed to get chaincode proposal payload input")
	}
	chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(cppInput)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal chaincodeInvocationSpec: %s", err)
	}
	chaincodeSpec := chaincodeInvocationSpec.GetChaincodeSpec()
	if chaincodeSpec == nil {
		return nil, fmt.Errorf("failed to get chaincode spec")
	}
	input := chaincodeSpec.GetInput()
	if input == nil {
		return nil, fmt.Errorf("failed to get chaincode spec input")
	}
	args := input.GetArgs()
	if args == nil {
		return nil, fmt.Errorf("failed to get chaincode spec input args")
	}

	// validate args
	// there two args:
	// 1. the function name (usually "__invoke") and
	// 2. the b64-encoded chaincode request message bytes
	if len(args) != 2 {
		return nil, fmt.Errorf("unexpected args num %d instead of 2 (function + chaincode request message", len(args))
	}

	// Return the decoded second arg
	chaincodeRequestMessageBytes, err := base64.StdEncoding.DecodeString(string(args[1]))
	if err != nil {
		return nil, fmt.Errorf("failed to decode chaincode request message")
	}
	return chaincodeRequestMessageBytes, nil
}

// UnwrapResponse unmarshalls the given serialized peer.Response message and returns the Payload field if Status is 200;
// otherwise, the Message field is returned as an error
func UnwrapResponse(responseBytes []byte) (payload []byte, err error) {
	clearResponse, err := protoutil.UnmarshalResponse(responseBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal peer.Response message")
	}

	if clearResponse.Status != 200 {
		return nil, errors.New(clearResponse.Message)
	}

	return clearResponse.Payload, nil
}
