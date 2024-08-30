/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/crypto"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/proto"
)

func experimentEndpoint(ep string) string {
	host := os.Getenv("EXPERIMENT_HOST")
	port := os.Getenv("EXPERIMENT_PORT")
	return fmt.Sprintf("http://%s:%s/%s", host, port, ep)
}

func requestAttestation() ([]byte, error) {
	resp, err := http.Get(experimentEndpoint("attestation"))
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("got: %s\n", bodyBytes)

	workerCredential := &pb.WorkerCredentials{}
	err = proto.Unmarshal(bodyBytes, workerCredential)
	if err != nil {
		return nil, err
	}

	fmt.Printf("got %s\n", workerCredential)

	identity := &pb.Identity{}
	err = proto.Unmarshal(workerCredential.GetIdentityBytes(), identity)
	if err != nil {
		return nil, err
	}

	fmt.Printf("got id: %s\n", identity.GetPublicEncryptionKey())

	return identity.GetPublicEncryptionKey(), nil
}

func submitEvaluationPack(pk []byte, req *pb.RegisterDataRequest) error {
	epm := &pb.EvaluationPackMessage{}
	epm.RegisteredData = []*pb.RegisterDataRequest{req}

	epmBytes, err := proto.Marshal(epm)
	if err != nil {
		return err
	}

	c := crypto.NewGoCrypto()

	k, err := c.NewSymmetricKey()
	if err != nil {
		return err
	}

	encryptedKey, err := c.PkEncryptMessage(pk, k)
	if err != nil {
		return err
	}

	encryptedEvaluationPack, err := c.EncryptMessage(k, epmBytes)
	if err != nil {
		return err
	}

	evalPack := &pb.EncryptedEvaluationPack{}
	evalPack.EncryptedEncryptionKey = encryptedKey
	evalPack.EncryptedEvaluationpack = encryptedEvaluationPack

	evalPackBytes, err := proto.Marshal(evalPack)
	if err != nil {
		return err
	}

	resp, err := http.Post(experimentEndpoint("execute-evaluationpack"), "", bytes.NewBuffer(evalPackBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Result: %s\n", bodyBytes)

	return nil
}

func upload() (*pb.RegisterDataRequest, error) {
	cp := crypto.NewGoCrypto()
	sk, err := cp.NewSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new symmetric key")
	}

	data := []byte("37.7, 0, 0, 1, 1, 0")

	encryptedData, err := cp.EncryptMessage(sk, data)
	if err != nil {
		return nil, errors.Wrap(err, "cannot encrypt message")
	}

	host := cmp.Or(os.Getenv("REDIS_HOST"), "localhost")
	port, err := strconv.Atoi(cmp.Or(os.Getenv("REDIS_PORT"), strconv.Itoa(storage.DefaultRedisPort)))
	if err != nil {
		return nil, errors.Wrap(err, "invalid redis port")
	}
	password := cmp.Or(os.Getenv("REDIS_PASSWORD"))

	// upload encrypted data
	kvs := storage.NewClient(storage.WithHost(host), storage.WithPort(port), storage.WithPassword(password))
	handle, err := kvs.Upload(encryptedData)
	if err != nil {
		return nil, errors.Wrap(err, "cannot upload data to kvs")
	}

	userIdentity := pb.Identity{
		Uuid:      "somePatient",
		PublicKey: []byte("some verification key"),
	}

	// build request
	registerDataRequest := &pb.RegisterDataRequest{
		Participant:   &userIdentity,
		DecryptionKey: sk,
		DataHandler:   handle,
		StudyId:       "some study",
	}

	return registerDataRequest, nil
}

const networkID = "mytestnetwork"

func TestWorker(t *testing.T) {
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
	redisImageName := "redis"
	redisExportedPort := fmt.Sprintf("%d/tcp", storage.DefaultRedisPort)
	redis, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        redisImageName,
			ExposedPorts: []string{redisExportedPort},
			Networks:     []string{networkName},
			// WaitingFor:   wait.ForLog("* Ready to accept connections"),
			WaitingFor: wait.ForExposedPort(),
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
	redisPort, err := redis.MappedPort(ctx, nat.Port(redisExportedPort))
	require.NoError(t, err)
	t.Logf("redisHost: %s:%s", redisHost, redisPort.Port())

	os.Setenv("REDIS_HOST", redisHost)
	os.Setenv("REDIS_PORT", redisPort.Port())

	// experiment
	experimentImageName := "irb-experimenter-worker"
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

	experimentHost, err := experiment.Host(ctx)
	require.NoError(t, err)
	experimentPort, err := experiment.MappedPort(ctx, "5000/tcp")
	require.NoError(t, err)
	t.Logf("experimentHost: %s:%s", experimentHost, experimentPort.Port())

	os.Setenv("EXPERIMENT_HOST", experimentHost)
	os.Setenv("EXPERIMENT_PORT", experimentPort.Port())

	req, err := upload()
	require.NoError(t, err)

	fmt.Println("Testing attestation...")
	pk, err := requestAttestation()
	require.NoError(t, err)

	fmt.Println("Testing evaluation pack...")
	err = submitEvaluationPack(pk, req)
	require.NoError(t, err)

	fmt.Println("Test done.")
}
