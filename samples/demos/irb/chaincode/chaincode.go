package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"google.golang.org/protobuf/proto"
)

type Asset struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type SmartContract struct {
	contractapi.Contract
}

func (s *SmartContract) RegisterData(ctx contractapi.TransactionContextInterface, registerDataRequestBase64 string) error {

	log.Printf("Process RegisterData")

	registerDataRequestBytes, err := base64.StdEncoding.DecodeString(registerDataRequestBase64)
	if err != nil {
		return fmt.Errorf("RegisterDataRequest base64 decoding failed: %v", err)
	}

	registerDataRequest := pb.RegisterDataRequest{}
	if err := proto.Unmarshal(registerDataRequestBytes, &registerDataRequest); err != nil {
		return fmt.Errorf("RegisterDataRequest unmarshalling failed: %v", err)
	}

	if err := ctx.GetStub().PutState("user."+registerDataRequest.Participant.Uuid+".uuid", []byte(registerDataRequest.Participant.Uuid)); err != nil {
		return fmt.Errorf("failed to put patient uuid: %v", err)
	}

	if err := ctx.GetStub().PutState("user."+registerDataRequest.Participant.Uuid+".pk", registerDataRequest.Participant.PublicKey); err != nil {
		return fmt.Errorf("failed to put patient vk: %v", err)
	}

	if err := ctx.GetStub().PutState("user."+registerDataRequest.Participant.Uuid+".data.handler", []byte(registerDataRequest.DataHandler)); err != nil {
		return fmt.Errorf("failed to put data handler: %v", err)
	}

	if err := ctx.GetStub().PutState("user."+registerDataRequest.Participant.Uuid+".data.dk", []byte(base64.StdEncoding.EncodeToString(registerDataRequest.DecryptionKey))); err != nil {
		return fmt.Errorf("failed to put decryption: %v", err)
	}

	return nil
}

func (s *SmartContract) NewExperiment(ctx contractapi.TransactionContextInterface, experimentProposalB64 string) error {

	log.Printf("Process NewExperiment")

	experimentProposalBytes, err := base64.StdEncoding.DecodeString(experimentProposalB64)
	if err != nil {
		return fmt.Errorf("ExperimentProposal base64 decoding failed: %v", err)
	}

	experimentProposal := pb.ExperimentProposal{}
	if err := proto.Unmarshal(experimentProposalBytes, &experimentProposal); err != nil {
		return fmt.Errorf("ExperimentProposal unmarshalling failed: %v", err)
	}

	// check if already exists
	d, err := ctx.GetStub().GetState("experimenter." + experimentProposal.GetExperimentId())
	if err != nil {
		return fmt.Errorf("failed to get experimenter: %v", err)
	}

	if d != nil {
		return fmt.Errorf("experimenter with id=%s already registered", experimentProposal.GetExperimentId())
	}

	// store
	if err := ctx.GetStub().PutState("experimenter."+experimentProposal.GetExperimentId(), []byte(experimentProposalB64)); err != nil {
		return fmt.Errorf("failed to put experimenter proposal: %v", err)
	}

	return nil
}

func (s *SmartContract) RegisterStudy(ctx contractapi.TransactionContextInterface, studyDetailsMessageBytesB64 string) error {

	log.Printf("Process RegisterStudy")

	studyDetailsMessageBytes, err := base64.StdEncoding.DecodeString(studyDetailsMessageBytesB64)
	if err != nil {
		return fmt.Errorf("StudyDetailsMessage base64 decoding failed: %v", err)
	}

	studyDetailsMessage := pb.StudyDetailsMessage{}
	if err := proto.Unmarshal(studyDetailsMessageBytes, &studyDetailsMessage); err != nil {
		return fmt.Errorf("StudyDetailsMessage unmarshalling failed: %v", err)
	}

	// check if already exists
	d, err := ctx.GetStub().GetState("study." + studyDetailsMessage.GetStudyId())
	if err != nil {
		return fmt.Errorf("failed to get study: %v", err)
	}

	if d != nil {
		return fmt.Errorf("study with id=%s already registered", studyDetailsMessage.GetStudyId())
	}

	// store
	if err := ctx.GetStub().PutState("study."+studyDetailsMessage.GetStudyId(), []byte(studyDetailsMessageBytesB64)); err != nil {
		return fmt.Errorf("failed to put study proposal: %v", err)
	}

	return nil
}

func (s *SmartContract) GetExperimentProposal(ctx contractapi.TransactionContextInterface, getExperimentRequestByteB64 string) (string, error) {

	log.Printf("Process GetExperimentProposal")

	getExperimentRequestByte, err := base64.StdEncoding.DecodeString(getExperimentRequestByteB64)
	if err != nil {
		return "", fmt.Errorf("GetExperimentRequest base64 decoding failed: %v", err)
	}

	getExperimentRequest := pb.GetExperimentRequest{}
	if err := proto.Unmarshal(getExperimentRequestByte, &getExperimentRequest); err != nil {
		return "", fmt.Errorf("GetExperimentRequest unmarshalling failed: %v", err)
	}

	// check if already exists
	d, err := ctx.GetStub().GetState("experimenter." + getExperimentRequest.GetExperimentId())
	if err != nil {
		return "", fmt.Errorf("failed to get experimenter: %v", err)
	}

	return string(d), nil
}

func (s *SmartContract) ApproveExperiment(ctx contractapi.TransactionContextInterface, signedApprovalBytesB64 string) error {

	log.Printf("Process ApproveExperiment")

	signedApprovalBytes, err := base64.StdEncoding.DecodeString(signedApprovalBytesB64)
	if err != nil {
		return fmt.Errorf("signedApproval base64 decoding failed: %v", err)
	}

	signedApproval := pb.SignedApprovalMessage{}
	if err := proto.Unmarshal(signedApprovalBytes, &signedApproval); err != nil {
		return fmt.Errorf("signedApproval unmarshalling failed: %v", err)
	}

	approval := pb.Approval{}
	if err := proto.Unmarshal(signedApproval.GetApproval(), &approval); err != nil {
		return fmt.Errorf("signedApproval unmarshalling failed: %v", err)
	}

	// get experimenter
	d, err := ctx.GetStub().GetState("experimenter." + approval.GetExperimentId())
	if err != nil {
		return fmt.Errorf("failed to get experimenter: %v", err)
	}

	if d == nil {
		return fmt.Errorf("experimenter with id=%s does not exist", approval.GetExperimentId())
	}

	// store approval
	if err := ctx.GetStub().PutState("approval."+approval.GetExperimentId(), []byte(signedApprovalBytesB64)); err != nil {
		return fmt.Errorf("failed to put study proposal: %v", err)
	}

	return nil
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetName string, assetValue int) error {

	log.Printf("Create asset with name '%s' and value '%d'", assetName, assetValue)

	asset := Asset{
		assetName,
		assetValue,
	}
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to create asset JSON: %v", err)
	}

	err = ctx.GetStub().PutState(asset.Name, assetBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset in public data: %v", err)
	}

	return nil
}

func (s *SmartContract) RetrieveAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {
	// Since only public data is accessed in this function, no access control is required
	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("%s does not exist", assetID)
	}

	var asset *Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}
	return asset, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panicf("Error create transfer asset chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting asset chaincode: %v", err)
	}
}
