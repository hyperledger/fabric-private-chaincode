/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package investigator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/fpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/messages"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type ApprovalView struct {
}

func (c *ApprovalView) Call(context view.Context) (interface{}, error) {
	session := context.Session()
	ch := session.Receive()

	msg := &messages.ApprovalRequestNotification{}
	select {
	case m := <-ch:
		if err := json.Unmarshal(m.Payload, msg); err != nil {
			return nil, err
		}
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("time out reached")
	}
	fmt.Printf("Received approval request from: %s\n%s\n", msg.Sender, msg.Message)

	if err := c.reviewAndApprove(context, msg.ExperimentID); err != nil {
		msg := fmt.Sprintf("something went wrong, cannot approve, sorry! reason: %v", err)
		_ = session.SendError([]byte(msg))
		return nil, errors.Wrap(err, msg)
	}

	if err := session.Send([]byte("great experimenter! approved!")); err != nil {
		return nil, errors.Wrap(err, "cannot send reply")
	}

	return nil, nil
}

func (c *ApprovalView) reviewAndApprove(context view.Context, experimentID string) error {
	fmt.Println("Trying to get experiment proposal")

	cid := "experimenter-approval-service"
	f := "getExperimentProposal"
	arg := utils.MarshalProtoBase64(&pb.GetExperimentRequest{
		ExperimentId: experimentID,
	})

	res, err := fpc.GetDefaultChannel(context).Chaincode(cid).Query(f, arg).Call()
	if err != nil {
		return errors.Wrapf(err, "error invoking %s", f)
	}

	experimentProposalBytes, err := base64.StdEncoding.DecodeString(string(res))
	if err != nil {
		return errors.Wrap(err, "cannot decode experiment proposal bytes")
	}

	experimentProposal := &pb.ExperimentProposal{}
	if err = proto.Unmarshal(experimentProposalBytes, experimentProposal); err != nil {
		return errors.Wrap(err, "cannot unmarshal experiment proposal")
	}

	// review
	// TODO review experimentProposal

	// make a decision
	approvalDecision := pb.Approval_APPROVED
	//approvalDecision := pb.Approval_UNDEFINED
	//approvalDecision = pb.Approval_REJECTED

	approval := pb.Approval{
		ExperimentId:       experimentProposal.GetExperimentId(),
		ExperimentProposal: experimentProposalBytes,
		Decision:           approvalDecision,
	}

	approvalBytes, err := proto.Marshal(&approval)
	if err != nil {
		return errors.Wrap(err, "cannot marshal approval")
	}

	f = "approveExperiment"
	arg = utils.MarshalProtoBase64(&pb.SignedApprovalMessage{
		Approval: approvalBytes,
		// TODO create approval signature; note that the chaincode currently does not check the signature
		//Signature: signature,
	})

	if _, err := fpc.GetDefaultChannel(context).Chaincode(cid).Invoke(f, arg).Call(); err != nil {
		return errors.Wrapf(err, "error invoking %s", f)
	}

	fmt.Println("LGTM! Approved!")
	return nil
}
