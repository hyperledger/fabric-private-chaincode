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

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

func MarshallProto(msg proto.Message) string {
	return base64.StdEncoding.EncodeToString(protoutil.MarshalOrPanic(msg))
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

// returns enclave_id as hex-encoded string of SHA256 hash over enclave_vk.
func GetEnclaveId(attestedData *protos.AttestedData) string {
	// hash enclave vk
	h := sha256.Sum256(attestedData.EnclaveVk)
	// encode and normalize
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

func ExtractEndpoint(credentials *protos.Credentials) (string, error) {
	attestedData := &protos.AttestedData{}
	err := ptypes.UnmarshalAny(credentials.SerializedAttestedData, attestedData)
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

	proposal := &pb.Proposal{}
	err = proto.Unmarshal(signedProposal.ProposalBytes, proposal)
	if err != nil {
		return nil, fmt.Errorf("failed to extract Proposal from SignedProposal: %s", err)
	}

	// check for header
	if len(proposal.GetHeader()) == 0 {
		return nil, errors.New("failed to extract Proposal fields: proposal header is nil")
	}

	// extract header
	hdr := &common.Header{}
	if err := proto.Unmarshal(proposal.GetHeader(), hdr); err != nil {
		return nil, fmt.Errorf("failed to extract proposal header: %s", err)
	}

	// validate channel header
	chdr := &common.ChannelHeader{}
	if err := proto.Unmarshal(hdr.ChannelHeader, chdr); err != nil {
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
	payload := &pb.ChaincodeProposalPayload{}
	if err := proto.Unmarshal(proposal.GetPayload(), payload); err != nil {
		return nil, fmt.Errorf("failed to extract proposal payload: %s", err)
	}
	cppInput := payload.GetInput()
	if cppInput == nil {
		return nil, fmt.Errorf("failed to get chaincode proposal payload input")
	}
	chaincodeInvocationSpec := &pb.ChaincodeInvocationSpec{}
	err = proto.Unmarshal(cppInput, chaincodeInvocationSpec)
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
