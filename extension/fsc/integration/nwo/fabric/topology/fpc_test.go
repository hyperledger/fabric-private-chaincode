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
