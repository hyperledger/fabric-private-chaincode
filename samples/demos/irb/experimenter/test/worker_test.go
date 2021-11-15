/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/container"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/crypto"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/storage"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/utils"
	"github.com/pkg/errors"
)

func requestAttestation() ([]byte, error) {

	resp, err := http.Get("http://localhost:5000/attestation")
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
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

	resp, err := http.Post("http://localhost:5000/execute-evaluationpack", "", bytes.NewBuffer(evalPackBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
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

	// upload encrypted data
	kvs := storage.NewClient()
	handle, err := kvs.Upload(encryptedData)
	if err != nil {
		return nil, errors.Wrap(err, "cannot upload data to kvs")
	}

	userIdentity := pb.Identity{
		Uuid:      "somePatient",
		PublicKey: []byte("some verification key"),
	}

	//build request
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

	network := &container.Network{Name: networkID}
	err := network.Create()
	defer network.Remove()
	if err != nil {
		panic(err)
	}

	// setup redis
	redis := &container.Container{
		Image:    "redis",
		Name:     "redis-container",
		HostIP:   "localhost",
		HostPort: "6379",
		Network:  networkID,
	}
	err = redis.Start()
	defer redis.Stop()
	if err != nil {
		panic(err)
	}

	// setup experiment container
	experiment := &container.Container{
		Image:    "irb-experimenter-worker",
		Name:     "experiment-container",
		HostIP:   "localhost",
		HostPort: "5000",
		Env:      []string{"REDIS_HOST=redis-container"},
		Network:  networkID,
	}
	err = experiment.Start()
	defer experiment.Stop()
	if err != nil {
		panic(err)
	}

	// let's wait until experiment service is up an running
	err = utils.Retry(func() bool {
		resp, err := http.Get("http://localhost:5000/info")
		if err != nil {
			return false
		}
		return resp.StatusCode == 200
	}, 5, 60*time.Second, 2*time.Second)
	if err != nil {
		panic(err)
	}

	req, err := upload()
	if err != nil {
		panic(err)
	}

	fmt.Println("Testing attestation...")
	pk, err := requestAttestation()
	if err != nil {
		panic(err)
	}

	fmt.Println("Testing evaluation pack...")
	if err := submitEvaluationPack(pk, req); err != nil {
		panic(err)
	}

	fmt.Println("Test done.")
}
