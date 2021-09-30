package investigator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/chaincode"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/utils"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type ApprovalView struct {
}

func (c *ApprovalView) Call(context view.Context) (interface{}, error) {
	session := context.Session()
	ch := session.Receive()

	msg := &views.ApprovalRequestNotification{}
	select {
	case m := <-ch:
		if err := json.Unmarshal(m.Payload, msg); err != nil {
			return nil, err
		}
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("time out reached")
	}

	fmt.Printf("Received msg from: %s\n%s\n", msg.Sender, msg.Message)

	// get experimenter proposal
	getExperimentRequest := &pb.GetExperimentRequest{
		ExperimentId: msg.ExperimentID,
	}

	response, err := context.RunView(
		chaincode.NewInvokeView(
			"experimenter-approval-service",
			"GetExperimentProposal",
			utils.MarshalProtoBase64(getExperimentRequest),
		).WithEndorsersFromMyOrg(),
	)
	if err != nil {
		// tell experimenter that something went wrong with approving
		msg := fmt.Sprintf("something went wrong, cannot approve; reason: %v", err)
		_ = session.SendError([]byte(msg))
		return nil, errors.Wrap(err, "error invoking chaincode")
	}

	_, res, err := unboxChaincodeResponse(response)
	if err != nil {
		return nil, err
	}

	experimentProposalBytes, err := base64.StdEncoding.DecodeString(string(res))
	if err != nil {
		return nil, err
	}

	experimentProposal := &pb.ExperimentProposal{}
	if err = proto.Unmarshal(experimentProposalBytes, experimentProposal); err != nil {
		return nil, err
	}

	if err := utils.UnmarshalProtoBase64(res, experimentProposal); err != nil {
		return nil, err
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
		return nil, err
	}

	signedApproval := &pb.SignedApprovalMessage{
		Approval: approvalBytes,
		//Signature: signature,
	}

	// approve experimenter
	if _, err = context.RunView(
		chaincode.NewInvokeView(
			"experimenter-approval-service",
			"ApproveExperiment",
			utils.MarshalProtoBase64(signedApproval),
		).WithEndorsersFromMyOrg(),
	); err != nil {
		// tell experiemnter that something went wrong with approving
		msg := fmt.Sprintf("something went wrong, cannot approve; reason: %v", err)
		_ = session.SendError([]byte(msg))
		return nil, errors.Wrap(err, "error invoking chaincode")
	}

	if err := session.Send([]byte("great experimenter! approved!")); err != nil {
		return nil, errors.Wrap(err, "cannot send reply")
	}

	return nil, nil
}

func unboxChaincodeResponse(response interface{}) (txid string, result []byte, err error) {
	s := reflect.ValueOf(response)
	if s.Kind() != reflect.Slice {
		return "", nil, fmt.Errorf("InterfaceSlice() given a non-slice type")
	}
	if s.IsNil() {
		return "", nil, nil
	}
	if s.Len() != 2 {
		return "", nil, fmt.Errorf("expected lengh is two")
	}

	txid, ok := s.Index(0).Interface().(string)
	if !ok {
		return "", nil, fmt.Errorf("first arg not a string")
	}

	b, ok := s.Index(1).Interface().([]byte)
	if !ok {
		return "", nil, fmt.Errorf("second arg not []byte")
	}

	return txid, b, nil
}
