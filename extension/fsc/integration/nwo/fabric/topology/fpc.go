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

// extraParamsKey is the key of the FPC chaincodes in topology.Topology.ExtraParams
const extraParamsKey = "fpc.chaincodes"

// Chaincode is a Fabric Private Chaincode. FPC chaincodes are not added to topology.Topology.Chaincodes,
// as FSC deploys every chaincode listed there itself; the FPC NWO extension deploys them instead.
type Chaincode struct {
	topology.ChannelChaincode
	// Image is the docker image of the chaincode, which runs as an external chaincode
	Image string
	// SGXMode is either SIM or HW
	SGXMode string
	// SGXDevicesPaths are the SGX devices to attach to the chaincode container in HW mode
	SGXDevicesPaths []string
}

// Chaincodes returns the FPC chaincodes of t
func Chaincodes(t *topology.Topology) []*Chaincode {
	ccs, _ := t.ExtraParams[extraParamsKey].([]*Chaincode)
	return ccs
}

// EnableFPC registers the Enclave Registry Chaincode (ERCC) on the topology.
// Safe to call multiple times; ERCC is registered only once.
func EnableFPC(t *topology.Topology) {
	for _, cc := range Chaincodes(t) {
		if cc.Chaincode.Name == "ercc" {
			return
		}
	}
	AddFPCAtOrgs(t, "ercc", "fpc/ercc", nil)
}

// AddFPCAtOrgs adds a Fabric Private Chaincode to t, installing only on the listed orgs.
// If orgs is empty, installs on all registered organizations.
// The endorsement policy is set to majority of the organizations on which the chaincode is installed.
func AddFPCAtOrgs(t *topology.Topology, name, image string, orgs []string, options ...func(*Chaincode)) *Chaincode {
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

	cc := &Chaincode{
		ChannelChaincode: topology.ChannelChaincode{
			Chaincode: topology.Chaincode{
				Name:            name,
				Version:         "Version-1.0",
				Sequence:        "1",
				InitRequired:    false,
				Path:            name,
				Lang:            "external", // FPC CC runs as a separate docker container (Doesnt't run inside Fabric peer process)
				Label:           fmt.Sprintf("%s_1.0", name),
				Ctor:            `{"Args":["init"]}`,
				SignaturePolicy: policy,
			},
			Channel: t.Channels[0].Name,
			Peers:   peers,
		},
		Image:   image,
		SGXMode: "sim",
	}

	for _, o := range options {
		o(cc)
	}

	if t.ExtraParams == nil {
		t.ExtraParams = map[string]any{}
	}
	t.ExtraParams[extraParamsKey] = append(Chaincodes(t), cc)

	return cc
}

// AddFPC adds a Fabric Private Chaincode to t, installing on all registered organizations.
// The endorsement policy is set to majority of the organizations on which the chaincode is installed.
func AddFPC(t *topology.Topology, name, image string, options ...func(*Chaincode)) *Chaincode {
	return AddFPCAtOrgs(t, name, image, nil, options...)
}

// WithSGXMode sets the SGX mode (HW or SIM) for the chaincode.
func WithSGXMode(mode string) func(*Chaincode) {
	return func(cc *Chaincode) {
		cc.SGXMode = mode
	}
}

// WithSGXDevicesPaths sets the SGX device paths for hardware mode.
func WithSGXDevicesPaths(paths []string) func(*Chaincode) {
	return func(cc *Chaincode) {
		cc.SGXDevicesPaths = paths
	}
}

// WithMREnclave sets the MRENCLAVE measurement as the chaincode version.
func WithMREnclave(mrenclave string) func(*Chaincode) {
	return func(cc *Chaincode) {
		cc.Chaincode.Version = mrenclave
	}
}
