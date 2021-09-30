package dataprovider

import (
	"encoding/json"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/chaincode"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/crypto"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/storage"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/utils"
	"github.com/pkg/errors"
)

type Register struct {
	StudyID     string
	PatientData []byte
	PatientUUID string
	PatientVK   []byte
}

type RegisterView struct {
	*Register
}

func (c *RegisterView) Call(context view.Context) (interface{}, error) {
	// encrypt with new random key
	cp := crypto.NewGoCrypto()
	sk, err := cp.NewSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new symmetric key")
	}

	encryptedData, err := cp.EncryptMessage(sk, c.PatientData)
	if err != nil {
		return nil, errors.Wrap(err, "cannot encrypt message")
	}

	// upload encrypted data
	kvs := storage.NewClient()
	handle, err := kvs.Upload(encryptedData)
	if err != nil {
		return nil, errors.Wrap(err, "cannot upload data to kvs")
	}

	userIdentity := pb.Identity{
		Uuid:      c.PatientUUID,
		PublicKey: c.PatientVK,
	}

	//build request
	registerDataRequest := &pb.RegisterDataRequest{
		Participant:   &userIdentity,
		DecryptionKey: sk,
		DataHandler:   handle,
		StudyId:       c.StudyID,
	}

	if _, err := context.RunView(
		chaincode.NewInvokeView(
			"experimenter-approval-service",
			"RegisterData",
			utils.MarshalProtoBase64(registerDataRequest),
		).WithEndorsersFromMyOrg(),
	); err != nil {
		return nil, errors.Wrap(err, "error invoking chaincode")
	}

	return nil, nil
}

type RegisterViewFactory struct{}

func (c *RegisterViewFactory) NewView(in []byte) (view.View, error) {
	f := &RegisterView{Register: &Register{}}
	if err := json.Unmarshal(in, f.Register); err != nil {
		return nil, err
	}
	return f, nil
}
