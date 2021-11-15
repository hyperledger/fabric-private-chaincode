/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package experimenter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/fpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/experiment"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/messages"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/utils"
	"github.com/pkg/errors"
)

type SubmitExperiment struct {
	StudyId      string
	ExperimentId string
	Investigator view.Identity
}

type SubmitExperimentView struct {
	*SubmitExperiment
}

func (c *SubmitExperimentView) Call(context view.Context) (interface{}, error) {
	// get worker credentials
	workerCredentials, err := experiment.GetWorkerCredentials()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}

	// submit new experimenter proposal
	experimentProposal := &pb.ExperimentProposal{
		StudyId:           c.StudyId,
		ExperimentId:      c.ExperimentId,
		WorkerCredentials: workerCredentials,
	}
	cid := "experimenter-approval-service"
	f := "newExperiment"
	arg := utils.MarshalProtoBase64(experimentProposal)

	if _, err := fpc.GetDefaultChannel(context).Chaincode(cid).Invoke(f, arg).Call(); err != nil {
		return nil, errors.Wrap(err, "error invoking "+f)
	}

	// reach out to investigator to trigger review and approval
	if err := c.waitForApprovals(context); err != nil {
		return nil, err
	}

	fmt.Println("It seems I got my approval! Ready to start the real work!")

	// trigger execution flow
	return context.RunView(NewExecutionView(c.ExperimentId))
}

func (c *SubmitExperimentView) waitForApprovals(context view.Context) error {
	session, err := context.GetSession(context.Initiator(), c.Investigator)
	if err != nil {
		return err
	}

	msg := &messages.ApprovalRequestNotification{
		Message:      "Hey my friend, I just submitted a new experimenter! Please review and approve asap; Thanks",
		Sender:       context.Me().String(),
		ExperimentID: c.ExperimentId,
	}

	raw, err := msg.Serialize()
	if err != nil {
		return err
	}

	if err = session.Send(raw); err != nil {
		return err
	}

	// wait for response
	ch := session.Receive()
	select {
	case msg := <-ch:
		if msg.Status != view.OK {
			return fmt.Errorf("got error: %v", string(msg.Payload))
		}
		fmt.Printf("Got answer: %v\n", string(msg.Payload))
	case <-time.After(1 * time.Minute):
		return fmt.Errorf("responder didn't answer in time")
	}

	return nil
}

type SubmitExperimentViewFactory struct{}

func (c *SubmitExperimentViewFactory) NewView(in []byte) (view.View, error) {
	f := &SubmitExperimentView{SubmitExperiment: &SubmitExperiment{}}
	if err := json.Unmarshal(in, f.SubmitExperiment); err != nil {
		return nil, err
	}
	return f, nil
}
