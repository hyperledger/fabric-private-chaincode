/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package protoutil provides the small subset of Fabric's protoutil helpers
// that FPC needs. It is intentionally self-contained (rather than importing
// fabric-x-common/protoutil) so that FPC does not depend on protobuf helpers
// that may diverge from classic Fabric's wire format over time.
package protoutil

import (
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// Marshal serializes a protobuf message.
func Marshal(pb proto.Message) ([]byte, error) {
	if !pb.ProtoReflect().IsValid() {
		return nil, errors.New("proto: Marshal called with nil")
	}
	return proto.Marshal(pb)
}

// MarshalOrPanic serializes a protobuf message and panics if this
// operation fails.
func MarshalOrPanic(pb proto.Message) []byte {
	if !pb.ProtoReflect().IsValid() {
		panic(errors.New("proto: Marshal called with nil"))
	}
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

// UnmarshalProposal unmarshals bytes to a Proposal.
func UnmarshalProposal(propBytes []byte) (*peer.Proposal, error) {
	prop := &peer.Proposal{}
	err := proto.Unmarshal(propBytes, prop)
	return prop, errors.Wrap(err, "error unmarshalling Proposal")
}

// UnmarshalChaincodeProposalPayload unmarshals bytes to a ChaincodeProposalPayload.
func UnmarshalChaincodeProposalPayload(bytes []byte) (*peer.ChaincodeProposalPayload, error) {
	cpp := &peer.ChaincodeProposalPayload{}
	err := proto.Unmarshal(bytes, cpp)
	return cpp, errors.Wrap(err, "error unmarshalling ChaincodeProposalPayload")
}

// UnmarshalChaincodeInvocationSpec unmarshals bytes to a ChaincodeInvocationSpec.
func UnmarshalChaincodeInvocationSpec(encoded []byte) (*peer.ChaincodeInvocationSpec, error) {
	cis := &peer.ChaincodeInvocationSpec{}
	err := proto.Unmarshal(encoded, cis)
	return cis, errors.Wrap(err, "error unmarshalling ChaincodeInvocationSpec")
}

// UnmarshalHeader unmarshals bytes to a Header.
func UnmarshalHeader(bytes []byte) (*common.Header, error) {
	hdr := &common.Header{}
	err := proto.Unmarshal(bytes, hdr)
	return hdr, errors.Wrap(err, "error unmarshalling Header")
}

// UnmarshalChannelHeader unmarshals bytes to a ChannelHeader.
func UnmarshalChannelHeader(bytes []byte) (*common.ChannelHeader, error) {
	chdr := &common.ChannelHeader{}
	err := proto.Unmarshal(bytes, chdr)
	return chdr, errors.Wrap(err, "error unmarshalling ChannelHeader")
}

// UnmarshalSignatureHeader unmarshals bytes to a SignatureHeader.
func UnmarshalSignatureHeader(bytes []byte) (*common.SignatureHeader, error) {
	sh := &common.SignatureHeader{}
	err := proto.Unmarshal(bytes, sh)
	return sh, errors.Wrap(err, "error unmarshalling SignatureHeader")
}

// UnmarshalSerializedIdentity unmarshals bytes to a SerializedIdentity.
func UnmarshalSerializedIdentity(bytes []byte) (*msp.SerializedIdentity, error) {
	sid := &msp.SerializedIdentity{}
	err := proto.Unmarshal(bytes, sid)
	return sid, errors.Wrap(err, "error unmarshalling SerializedIdentity")
}
