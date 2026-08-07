/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/contract"

	"github.com/hyperledger-labs/fabric-smart-client/platform/common/services/logging"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric"
	services "github.com/hyperledger-labs/fabric-smart-client/platform/view/services"
)

var logger = logging.MustGetLogger("fabric-sdk.fpc")

// Channel models a Fabric channel that supports invocation of a Fabric Private Chaincode
type Channel struct {
	FabricNetworkService *fabric.NetworkService
	Channel              *fabric.Channel

	ER *EnclaveRegistry
}

func newChannel(fns *fabric.NetworkService, ch *fabric.Channel, er *EnclaveRegistry) *Channel {
	return &Channel{FabricNetworkService: fns, Channel: ch, ER: er}
}

// EnclaveRegistry returns the enclave registry for this channel
func (p *Channel) EnclaveRegistry() *EnclaveRegistry {
	return p.ER
}

// Chaincode returns a wrapper around the Fabric Private Chaincode whose name is cid
func (p *Channel) Chaincode(cid string) *Chaincode {
	icp := &contractProvider{
		fns: p.FabricNetworkService,
		ch:  p.Channel,
	}

	return NewChaincode(
		p.Channel,
		p.ER,
		fpc.GetContract(icp, cid),
		&endorserContractImpl{fns: p.FabricNetworkService, ch: p.Channel, cid: cid},
		p.FabricNetworkService.IdentityProvider().DefaultIdentity(),
		p.FabricNetworkService.IdentityProvider(),
		cid,
	)
}

// GetDefaultChannel returns the default channel on which to invoke an FPC.
// Panics if the network or channel cannot be resolved (integration/test use only).
func GetDefaultChannel(sp services.Provider) *Channel {
	fns, err := fabric.GetDefaultFNS(sp)
	if err != nil {
		panic(err)
	}
	ch, err := fns.Channel("")
	if err != nil {
		panic(err)
	}
	return newChannel(fns, ch, NewEnclaveRegistry(fns, ch))
}

// GetChannel returns the channel for the passed network and channel name on which to invoke an FPC.
// Panics if the network or channel cannot be resolved (integration/test use only).
func GetChannel(sp services.Provider, network, channelName string) *Channel {
	fns, err := fabric.GetFabricNetworkService(sp, network)
	if err != nil {
		panic(err)
	}
	ch, err := fns.Channel(channelName)
	if err != nil {
		panic(err)
	}
	return newChannel(fns, ch, NewEnclaveRegistry(fns, ch))
}
