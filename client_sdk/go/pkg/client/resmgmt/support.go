/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/lifecycle"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"github.com/pkg/errors"
)

// Gateway interface that is needed by the FPC resource management client.
// It is implemented by *client.Gateway of the Fabric Gateway client API.
type Gateway interface {
	Identity() identity.Identity
	GetNetwork(name string) *client.Network
}

type channelClient struct {
	network *client.Network
	sign    identity.Sign
	dial    gateway.PeerDialer
}

// Query evaluates a chaincode function directly at the target peer, as the Fabric Gateway routes requests
// by organization rather than by peer.
func (c *channelClient) Query(chaincodeID string, fcn string, args [][]byte, targetEndpoints ...string) ([]byte, error) {
	if len(targetEndpoints) != 1 {
		return nil, errors.Errorf("exactly one target peer endpoint required, got %d", len(targetEndpoints))
	}
	return gateway.EvaluateAtPeer(c.network.GetContract(chaincodeID), c.sign, c.dial, targetEndpoints[0], fcn, args...)
}

// Execute submits a chaincode transaction, waits for it to be committed, and returns its transaction ID.
func (c *channelClient) Execute(chaincodeID string, fcn string, args [][]byte) (string, error) {
	proposal, err := c.network.GetContract(chaincodeID).NewProposal(fcn, client.WithBytesArguments(args...))
	if err != nil {
		return "", err
	}

	transaction, err := proposal.Endorse()
	if err != nil {
		return "", gateway.WithErrorDetails(err)
	}

	commit, err := transaction.Submit()
	if err != nil {
		return "", gateway.WithErrorDetails(err)
	}

	status, err := commit.Status()
	if err != nil {
		return "", gateway.WithErrorDetails(err)
	}
	if !status.Successful {
		return "", errors.Errorf("transaction %s failed to commit with status code %d", status.TransactionID, int32(status.Code))
	}

	return transaction.TransactionID(), nil
}

type ChannelClientProvider struct {
	gateway Gateway
	sign    identity.Sign
	dial    gateway.PeerDialer
}

func NewChannelClientProvider(gw Gateway, sign identity.Sign, dial gateway.PeerDialer) *ChannelClientProvider {
	return &ChannelClientProvider{gateway: gw, sign: sign, dial: dial}
}

func (c *ChannelClientProvider) ChannelClient(id string) (lifecycle.ChannelClient, error) {
	return &channelClient{
		network: c.gateway.GetNetwork(id),
		sign:    c.sign,
		dial:    c.dial,
	}, nil
}
