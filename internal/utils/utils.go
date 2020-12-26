/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

const MrEnclaveStateKey = "MRENCLAVE"

// Response contains the response data and signature produced by the enclave
// TODO remove once ecc uses new ChaincodeResponseMessage
type Response struct {
	ResponseData []byte `json:"ResponseData"`
	Signature    []byte `json:"Signature"`
	PublicKey    []byte `json:"PublicKey"`
}

func UnmarshalResponse(respBytes []byte) (*Response, error) {
	response := &Response{}
	err := json.Unmarshal(respBytes, response)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling FPC response err: %s", err)
	}
	return response, nil
}

// TODO replace this with a proto? TBD
type AttestationParams struct {
	Params []string `json:"params"`
}

const sep = "."

func Read(file string) []byte {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	if data == nil {
		panic(fmt.Errorf("File is empty"))
	}
	return data
}

func IsFPCCompositeKey(comp string) bool {
	return strings.HasPrefix(comp, sep) && strings.HasSuffix(comp, sep)
}

func TransformToFPCKey(comp string) string {
	return strings.Replace(comp, "\x00", sep, -1)
}

func SplitFPCCompositeKey(comp_str string) []string {
	// check it has sep in front and end
	if !IsFPCCompositeKey(comp_str) {
		panic("comp_key has wrong format")
	}
	comp := strings.Split(comp_str, sep)
	return comp[1 : len(comp)-1]
}

func ValidateEndpoint(endpoint string) error {
	colon := strings.LastIndexByte(endpoint, ':')
	if colon == -1 {
		return fmt.Errorf("invalid format")
	}

	_, err := strconv.Atoi(endpoint[colon+1:])
	if err != nil {
		return errors.Wrap(err, "invalid port")
	}

	return nil
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
