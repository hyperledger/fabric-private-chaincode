/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"context"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	gatewaypb "github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// PeerDialer returns a gRPC connection to the peer at the given endpoint, as registered at the enclave registry.
// It is used to send requests directly to the peer hosting an enclave, as the Fabric Gateway routes requests
// by organization rather than by peer. The caller is responsible for choosing the TLS settings of each peer.
type PeerDialer func(endpoint string) (grpc.ClientConnInterface, error)

// EvaluateAtPeer evaluates a transaction directly at the peer with the given endpoint using the peer's endorser
// gRPC API, similar to the Fabric Admin SDK, and returns the chaincode response payload.
// The proposal is created by the Fabric Gateway contract and signed with sign, which must belong to the gateway identity.
func EvaluateAtPeer(c *client.Contract, sign identity.Sign, dial PeerDialer, endpoint string, name string, args ...[]byte) ([]byte, error) {
	signedProposal, err := newSignedProposal(c, sign, name, args...)
	if err != nil {
		return nil, err
	}

	conn, err := dial(endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to peer %s", endpoint)
	}

	resp, err := peer.NewEndorserClient(conn).ProcessProposal(context.Background(), signedProposal)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to evaluate %s at peer %s", name, endpoint)
	}
	if status := resp.GetResponse().GetStatus(); status < 200 || status >= 400 {
		return nil, errors.Errorf("failed to evaluate %s at peer %s: status %d: %s", name, endpoint, status, resp.GetResponse().GetMessage())
	}
	return resp.GetResponse().GetPayload(), nil
}

// newSignedProposal builds a proposal with the Fabric Gateway contract and signs its digest, as done by the
// Fabric Gateway for offline signing
func newSignedProposal(c *client.Contract, sign identity.Sign, name string, args ...[]byte) (*peer.SignedProposal, error) {
	proposal, err := c.NewProposal(name, client.WithBytesArguments(args...))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create proposal")
	}
	signature, err := sign(proposal.Digest())
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign proposal")
	}

	proposalBytes, err := proposal.Bytes()
	if err != nil {
		return nil, err
	}
	proposed := &gatewaypb.ProposedTransaction{}
	if err := proto.Unmarshal(proposalBytes, proposed); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal proposal")
	}
	return &peer.SignedProposal{ProposalBytes: proposed.GetProposal().GetProposalBytes(), Signature: signature}, nil
}
