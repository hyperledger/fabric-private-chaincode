/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package irb

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hyperledger-labs/fabric-smart-client/integration"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/common"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/container"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/storage"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/users"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/dataprovider"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/experimenter"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/investigator"
	"github.com/stretchr/testify/assert"
)

const (
	studyID                 = "pineapple"
	experimentID            = "exp001"
	dockerNetwork           = "irb-network"
	redisImageName          = "redis"
	redisContainerName      = "redis-container"
	experimentImageName     = "irb-experimenter-worker"
	experimentContainerName = "experiment-container"
	testPatientData         = "29, 0, 0, 1, 0, 0"
)

func TestFlow(t *testing.T) {

	//
	network := &container.Network{Name: dockerNetwork}
	err := network.Create()
	assert.NoError(t, err)
	defer network.Remove()

	// setup redis
	redis := &container.Container{
		Image:    redisImageName,
		Name:     redisContainerName,
		HostIP:   "localhost",
		HostPort: strconv.Itoa(storage.DefaultRedisPort),
		Network:  dockerNetwork,
	}
	err = redis.Start()
	assert.NoError(t, err)
	defer redis.Stop()

	// setup experiment container
	experiment := &container.Container{
		Image:    experimentImageName,
		Name:     experimentContainerName,
		HostIP:   "localhost",
		HostPort: "5000",
		Network:  dockerNetwork,
		Env:      []string{fmt.Sprintf("REDIS_HOST=%s", redisContainerName)},
	}
	err = experiment.Start()
	defer experiment.Stop()
	assert.NoError(t, err)

	// setup fabric network
	ii, err := integration.Generate(23000, Topology()...)
	assert.NoError(t, err)
	ii.Start()
	defer ii.Stop()

	// create test patients participating in our study
	patients := []string{"patient1"}
	var patientIdentities []*pb.Identity
	for _, p := range patients {
		uuid, _, vk, err := users.LoadOrCreateUser(p)
		assert.NoError(t, err)
		patientIdentities = append(patientIdentities, &pb.Identity{Uuid: string(uuid), PublicKey: vk})
	}
	assert.Equal(t, len(patients), len(patientIdentities))

	// create new study with our patients
	_, err = ii.Client("investigator").CallView("CreateStudy", common.JSONMarshall(&investigator.CreateStudy{
		StudyID:      studyID,
		Metadata:     "some fancy study",
		Participants: patientIdentities,
	}))
	assert.NoError(t, err)

	// data provider flow
	_, err = ii.Client("provider").CallView("RegisterData", common.JSONMarshall(&dataprovider.Register{
		StudyID:     studyID,
		PatientData: []byte(testPatientData),
		PatientUUID: patientIdentities[0].GetUuid(),
		PatientVK:   patientIdentities[0].GetPublicKey(),
	}))
	assert.NoError(t, err)

	// starting experimenter flow
	// this starts with the submission view which interacts with the investigator approval view
	// once the new experiment is approved, the execution view is triggered
	_, err = ii.Client("experimenter").CallView("SubmitExperiment", common.JSONMarshall(&experimenter.SubmitExperiment{
		StudyId:      studyID,
		ExperimentId: experimentID,
		Investigator: ii.Identity("investigator"),
	}))
	assert.NoError(t, err)

}
