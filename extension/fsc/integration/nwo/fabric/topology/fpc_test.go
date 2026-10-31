/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package topology

import (
	"testing"

	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/topology"
)

func TestAddFPCReplacesChaincodeWithSameName(t *testing.T) {
	top := &topology.Topology{
		Consortiums: []*topology.Consortium{{Organizations: []string{"Org1"}}},
		Channels:    []*topology.Channel{{Name: "testchannel"}},
		Peers:       []*topology.Peer{{Name: "peer0", Organization: "Org1"}},
	}

	AddFPC(top, "cc", "img1")
	AddFPC(top, "cc", "img2")

	var names []string
	for _, cc := range Chaincodes(top) {
		names = append(names, cc.Chaincode.Name)
	}
	if len(names) != 2 || names[0] != "ercc" || names[1] != "cc" {
		t.Fatalf("got chaincodes %v, want [ercc cc]", names)
	}
	if img := Chaincodes(top)[1].Image; img != "img2" {
		t.Fatalf("got image %q, want img2", img)
	}
}

func TestErccImage(t *testing.T) {
	t.Setenv("FPC_ERCC_IMAGE", "")
	if img := ErccImage(); img != "fpc/ercc" {
		t.Fatalf("got %q, want fpc/ercc", img)
	}

	t.Setenv("FPC_ERCC_IMAGE", "fpc/ercc-nosgx")
	top := &topology.Topology{
		Consortiums: []*topology.Consortium{{Organizations: []string{"Org1"}}},
		Channels:    []*topology.Channel{{Name: "testchannel"}},
		Peers:       []*topology.Peer{{Name: "peer0", Organization: "Org1"}},
	}
	EnableFPC(top)
	if img := Chaincodes(top)[0].Image; img != "fpc/ercc-nosgx" {
		t.Fatalf("got ercc image %q, want fpc/ercc-nosgx", img)
	}
}
