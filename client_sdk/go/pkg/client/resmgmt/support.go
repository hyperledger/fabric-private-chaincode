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
	mspID   string
}

// Query evaluates a chaincode function at the peers of the organization of the client identity.
// The target endpoints are not used for routing, as the Fabric Gateway routes requests by organization
// and prefers the peer it runs on; the gateway connection must therefore be established to the target peer.
func (c *channelClient) Query(chaincodeID string, fcn string, args [][]byte, targetEndpoints ...string) ([]byte, error) {
	result, err := c.network.GetContract(chaincodeID).Evaluate(fcn,
		client.WithBytesArguments(args...),
		client.WithEndorsingOrganizations(c.mspID),
	)
	return result, gateway.WithErrorDetails(err)
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
}

func NewChannelClientProvider(gateway Gateway) *ChannelClientProvider {
	return &ChannelClientProvider{gateway: gateway}
}

func (c *ChannelClientProvider) ChannelClient(id string) (lifecycle.ChannelClient, error) {
	return &channelClient{
		network: c.gateway.GetNetwork(id),
		mspID:   c.gateway.Identity().MspID(),
	}, nil
}
