package investigator

import (
	"encoding/json"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/chaincode"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/utils"
	"github.com/pkg/errors"
)

type RegisterStudy struct {
	StudyID  string
	Metadata string
}

type RegisterStudyView struct {
	*RegisterStudy
}

func (c *RegisterStudyView) Call(context view.Context) (interface{}, error) {
	// todo get user identities

	//build request
	studyDetailsMessage := &pb.StudyDetailsMessage{
		StudyId:  c.StudyID,
		Metadata: c.Metadata,
		//UserIdentities: userIdentities,
	}

	if _, err := context.RunView(
		chaincode.NewInvokeView(
			"experimenter-approval-service",
			"RegisterStudy",
			utils.MarshalProtoBase64(studyDetailsMessage),
		).WithEndorsersFromMyOrg(),
	); err != nil {
		return nil, errors.Wrap(err, "error invoking chaincode")
	}

	return nil, nil
}

type RegisterStudyViewFactory struct{}

func (c *RegisterStudyViewFactory) NewView(in []byte) (view.View, error) {
	f := &RegisterStudyView{RegisterStudy: &RegisterStudy{}}
	if err := json.Unmarshal(in, f.RegisterStudy); err != nil {
		return nil, err
	}
	return f, nil
}
