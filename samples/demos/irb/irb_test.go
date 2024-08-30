/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package irb

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/hyperledger-labs/fabric-smart-client/integration"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/common"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/storage"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/users"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/dataprovider"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/experimenter"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/investigator"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
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
	ctx := context.Background()

	// network
	net, err := network.New(ctx)
	require.NoError(t, err)
	defer func() {
		err := net.Remove(ctx)
		require.NoError(t, err)
	}()

	networkName := net.Name

	// redis
	redis, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        redisImageName,
			ExposedPorts: []string{fmt.Sprintf("%d/tcp", storage.DefaultRedisPort)},
			Networks:     []string{networkName},
			WaitingFor:   wait.ForLog("* Ready to accept connections"),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer func() {
		err := redis.Terminate(ctx)
		require.NoError(t, err)
	}()

	redisHost, err := redis.Host(ctx)
	require.NoError(t, err)
	redisPort, err := redis.MappedPort(ctx, nat.Port(fmt.Sprintf("%d/tcp", storage.DefaultRedisPort)))
	require.NoError(t, err)

	// experiment
	experiment, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        experimentImageName,
			ExposedPorts: []string{"5000/tcp"},
			Networks:     []string{networkName},
			Env: map[string]string{
				"REDIS_HOST": redisHost,
				"REDIS_PORT": redisPort.Port(),
			},
			WaitingFor: wait.ForExposedPort(),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer func() {
		err := experiment.Terminate(ctx)
		require.NoError(t, err)
	}()

	experimenterHost, err := experiment.Host(ctx)
	require.NoError(t, err)
	experimenterPort, err := experiment.MappedPort(ctx, "5000/tcp")
	require.NoError(t, err)

	// setup fabric network
	ii, err := integration.Generate(23000, false, Topology()...)
	require.NoError(t, err)
	ii.Start()
	defer func() {
		ii.Stop()
		// remove generated cmd folder
		err := os.RemoveAll("cmd")
		require.NoError(t, err)
	}()

	// create test patients participating in our study
	patients := []string{"patient1"}
	var patientIdentities []*pb.Identity
	for _, p := range patients {
		uuid, _, vk, err := users.LoadOrCreateUser(p)
		require.NoError(t, err)
		patientIdentities = append(patientIdentities, &pb.Identity{Uuid: string(uuid), PublicKey: vk})
	}
	require.Equal(t, len(patients), len(patientIdentities))

	// create new study with our patients
	_, err = ii.Client("investigator").CallView("CreateStudy", common.JSONMarshall(&investigator.CreateStudy{
		StudyID:      studyID,
		Metadata:     "some fancy study",
		Participants: patientIdentities,
	}))
	require.NoError(t, err)

	// data provider flow
	_, err = ii.Client("provider").CallView("RegisterData", common.JSONMarshall(&dataprovider.Register{
		StudyID:     studyID,
		PatientData: []byte(testPatientData),
		PatientUUID: patientIdentities[0].GetUuid(),
		PatientVK:   patientIdentities[0].GetPublicKey(),
		StorageHost: redisHost,
		StoragePort: redisPort.Int(),
	}))
	require.NoError(t, err)

	// starting experimenter flow
	// this starts with the submission view which interacts with the investigator approval view
	// once the new experiment is approved, the execution view is triggered
	_, err = ii.Client("experimenter").CallView("SubmitExperiment", common.JSONMarshall(&experimenter.SubmitExperiment{
		StudyID:        studyID,
		ExperimentID:   experimentID,
		WorkerEndpoint: fmt.Sprintf("%s:%s", experimenterHost, experimenterPort.Port()),
		Investigator:   ii.Identity("investigator"),
	}))
	require.NoError(t, err)
}
