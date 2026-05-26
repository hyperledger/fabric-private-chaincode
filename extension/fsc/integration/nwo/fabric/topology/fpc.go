/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package topology

import (
	"fmt"
	"strconv"

	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/topology"
)

// EnableFPC registers the Enclave Registry Chaincode (ERCC) on the topology.
// Does not set topology.FPC - that field is dead code after FSC PR #696.
// Safe to call multiple times; ERCC is registered only once.
func EnableFPC(t *topology.Topology) {
	for _, cc := range t.Chaincodes {
		if cc.Chaincode.Name == "ercc" {
			return
		}
	}
	AddFPCAtOrgs(t, "ercc", "fpc/ercc", nil)
}

// AddFPCAtOrgs adds a Fabric Private Chaincode to t, installing only on the listed orgs.
// If orgs is empty, installs on all registered organizations.
// The endorsement policy is set to majority of the organizations on which the chaincode is installed.
func AddFPCAtOrgs(t *topology.Topology, name, image string, orgs []string, options ...func(*topology.ChannelChaincode)) *topology.ChannelChaincode {
	// EnableFPC registers ercc; skip when we are adding ercc itself to avoid recursion.
	if name != "ercc" {
		EnableFPC(t)
	}

	if len(orgs) == 0 {
		orgs = t.Consortiums[0].Organizations
	}
	majority := len(orgs)/2 + 1
	policy := "OutOf(" + strconv.Itoa(majority) + ","
	for i, org := range orgs {
		if i > 0 {
			policy += ","
		}
		policy += "'" + org + "MSP.member'"
	}
	policy += ")"

	var peers []string
	for _, org := range orgs {
		for _, peer := range t.Peers {
			if peer.Organization == org {
				peers = append(peers, peer.Name)
				break
			}
		}
	}

	cc := &topology.ChannelChaincode{
		Chaincode: topology.Chaincode{
			Name:            name,
			Version:         "Version-1.0",
			Sequence:        "1",
			InitRequired:    false,
			Path:            name,
			Lang:            "external",
			Label:           fmt.Sprintf("%s_1.0", name),
			Ctor:            `{"Args":["init"]}`,
			Policy:          policy,
			SignaturePolicy: policy,
		},
		PrivateChaincode: topology.PrivateChaincode{
			Image:   image,
			SGXMode: "sim",
		},
		Channel: t.Channels[0].Name,
		Private: true,
		Peers:   peers,
	}

	for _, o := range options {
		o(cc)
	}

	t.AddChaincode(cc)

	return cc
}

// AddFPC adds a Fabric Private Chaincode to t, installing on all registered organizations.
// The endorsement policy is set to majority of the organizations on which the chaincode is installed.
func AddFPC(t *topology.Topology, name, image string, options ...func(*topology.ChannelChaincode)) *topology.ChannelChaincode {
	return AddFPCAtOrgs(t, name, image, nil, options...)
}

// WithSGXMode sets the SGX mode (HW or SIM) for the chaincode.
func WithSGXMode(mode string) func(*topology.ChannelChaincode) {
	return func(cc *topology.ChannelChaincode) {
		cc.PrivateChaincode.SGXMode = mode
	}
}

// WithSGXDevicesPaths sets the SGX device paths for hardware mode.
func WithSGXDevicesPaths(paths []string) func(*topology.ChannelChaincode) {
	return func(cc *topology.ChannelChaincode) {
		cc.PrivateChaincode.SGXDevicesPaths = paths
	}
}

// WithMREnclave sets the MRENCLAVE measurement as the chaincode version.
func WithMREnclave(mrenclave string) func(*topology.ChannelChaincode) {
	return func(cc *topology.ChannelChaincode) {
		cc.Chaincode.Version = mrenclave
	}
}
