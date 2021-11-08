/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package irb

import (
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/api"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fsc"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/dataprovider"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/experimenter"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/investigator"
)

func Topology() []api.Topology {
	fabricTopology := fabric.NewDefaultTopology()
	fabricTopology.AddOrganizationsByName("Org1", "Org2", "Org3")
	fabricTopology.AddFPC("experimenter-approval-service", "fpc/irb-experiment")
	fscTopology := fsc.NewTopology()

	// data provider
	providerNode := fscTopology.AddNodeByName("provider")
	providerNode.AddOptions(fabric.WithOrganization("Org1"))
	providerNode.RegisterViewFactory("RegisterData", &dataprovider.RegisterViewFactory{})

	// investigator
	investigatorNode := fscTopology.AddNodeByName("investigator")
	investigatorNode.AddOptions(fabric.WithOrganization("Org2"))
	investigatorNode.RegisterViewFactory("CreateStudy", &investigator.CreateStudyViewFactory{})
	investigatorNode.RegisterResponder(&investigator.ApprovalView{}, &experimenter.SubmitExperimentView{})

	//experimenter
	experimenterNode := fscTopology.AddNodeByName("experimenter")
	experimenterNode.AddOptions(fabric.WithOrganization("Org3"))
	experimenterNode.RegisterViewFactory("SubmitExperiment", &experimenter.SubmitExperimentViewFactory{})

	return []api.Topology{fabricTopology, fscTopology}
}
