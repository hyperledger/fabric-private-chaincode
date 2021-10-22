/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package investigator

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/fpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/utils"
	"github.com/pkg/errors"
)

type CreateStudy struct {
	StudyID      string
	Metadata     string
	Participants []*pb.Identity
}

type CreateStudyView struct {
	*CreateStudy
}

func (c *CreateStudyView) Call(context view.Context) (interface{}, error) {
	fmt.Println("Let's register first a new study ")

	//build request
	studyDetailsMessage := &pb.StudyDetailsMessage{
		StudyId:        c.StudyID,
		Metadata:       c.Metadata,
		UserIdentities: c.Participants,
	}

	// chaincode details
	cid := "experimenter-approval-service"
	f := "registerStudy"
	arg := utils.MarshalProtoBase64(studyDetailsMessage)

	_, err := fpc.GetDefaultChannel(context).Chaincode(cid).Invoke(f, arg).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking %s", f)
	}

	fmt.Println("Study created! thanks")
	return nil, nil
}

type CreateStudyViewFactory struct{}

func (c *CreateStudyViewFactory) NewView(in []byte) (view.View, error) {
	f := &CreateStudyView{CreateStudy: &CreateStudy{}}
	if err := json.Unmarshal(in, f.CreateStudy); err != nil {
		return nil, err
	}
	return f, nil
}
