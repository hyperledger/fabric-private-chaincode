package experimenter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/chaincode"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/utils"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views"
	"github.com/pkg/errors"
)

type Submission struct {
	StudyId      string
	ExperimentId string
	Investigator view.Identity
}

type SubmissionView struct {
	*Submission
}

func (c *SubmissionView) Call(context view.Context) (interface{}, error) {
	// TODO start experimenter container (worker)

	// get worker credentials
	//workerCredentials, err := worker.GetWorkerCredentials()
	//if err != nil {
	//	fmt.Printf("error: %v\n", err)
	//	return nil, err
	//}

	// submit new experimenter proposal
	experimentProposal := &pb.ExperimentProposal{
		StudyId:      c.StudyId,
		ExperimentId: c.ExperimentId,
		//WorkerCredentials: workerCredentials,
	}
	if err := c.submitProposal(context, experimentProposal); err != nil {
		return nil, err
	}

	// reach out to investigator to trigger review and approval
	if err := c.waitForApprovals(context); err != nil {
		return nil, err
	}

	// trigger execution flow
	return context.RunView(NewExecutionView(c.ExperimentId))
}

func (c *SubmissionView) submitProposal(context view.Context, experimentProposal *pb.ExperimentProposal) error {
	if _, err := context.RunView(
		chaincode.NewInvokeView(
			"experimenter-approval-service",
			"newExperiment",
			utils.MarshalProtoBase64(experimentProposal),
		).WithEndorsersFromMyOrg(),
	); err != nil {
		return errors.Wrap(err, "error invoking chaincode")
	}

	return nil
}

func (c *SubmissionView) waitForApprovals(context view.Context) error {
	session, err := context.GetSession(context.Initiator(), c.Investigator)
	if err != nil {
		return err
	}

	msg := &views.ApprovalRequestNotification{
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

type SubmissionViewFactory struct{}

func (c *SubmissionViewFactory) NewView(in []byte) (view.View, error) {
	f := &SubmissionView{Submission: &Submission{}}
	if err := json.Unmarshal(in, f.Submission); err != nil {
		return nil, err
	}
	return f, nil
}
