/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package experimenter

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/fpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/experiment"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type Execution struct {
	ExperimentId string
}

type ExecutionView struct {
	*Execution
}

func (c *ExecutionView) Call(context view.Context) (interface{}, error) {
	fmt.Println("All cool, now we can run our experimenter")

	//build experiment proposal
	evaluationPackRequest := pb.EvaluationPackRequest{
		ExperimentId: c.ExperimentId,
	}

	evaluationPackRequestBytes, err := proto.Marshal(&evaluationPackRequest)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal evaluation pack request")
	}

	cid := "experimenter-approval-service"
	f := "requestEvaluationPack"
	arg := base64.StdEncoding.EncodeToString(evaluationPackRequestBytes)

	response, err := fpc.GetDefaultChannel(context).Chaincode(cid).Query(f, arg).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking %s", f)
	}

	encryptedEvaluationPackBytes, err := base64.StdEncoding.DecodeString(string(response))
	if err != nil {
		return nil, err
	}

	encryptedEvaluationPack := &pb.EncryptedEvaluationPack{}
	err = proto.Unmarshal(encryptedEvaluationPackBytes, encryptedEvaluationPack)
	if err != nil || encryptedEvaluationPack.GetEncryptedEvaluationpack() == nil {
		//error decoding means something wrong with making the pack
		status, e := utils.UnmarshalStatus(encryptedEvaluationPackBytes)
		if e != nil {
			//cannot even unmarshal status, so just return the error
			return nil, err
		}

		//return error from status
		m := fmt.Sprintf("error getExperimentProposal: %s, %s", status.GetReturnCode(), status.GetMsg())
		return nil, errors.New(m)
	}
	fmt.Println("Received evaluation pack from FPC Experiment Approval Service!")

	// next, we send the eval pack to the worker
	// TODO double check that the worker can access redis
	resultBytes, err := experiment.ExecuteEvaluationPack(encryptedEvaluationPack)
	fmt.Printf("Result received from worker: \"%s\"\n", string(resultBytes))

	return nil, nil
}

func NewExecutionView(ExperimentID string) view.View {
	return &ExecutionView{
		Execution: &Execution{ExperimentId: ExperimentID},
	}
}
