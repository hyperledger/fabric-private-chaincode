/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2021 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package experiment

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"google.golang.org/protobuf/proto"
)

func toEvidence(attestation []byte) ([]byte, error) {
	fpcPath := os.Getenv("FPC_PATH")
	if fpcPath == "" {
		return nil, fmt.Errorf("FPC_PATH not set")
	}

	convertScript := filepath.Join(fpcPath, "common/crypto/attestation-api/conversion/attestation_to_evidence.sh")
	cmd := exec.Command(convertScript, string(attestation))

	if out, err := cmd.Output(); err != nil {
		return nil, err
	} else {
		return []byte(strings.TrimSuffix(string(out), "\n")), nil
	}
}

func GetWorkerCredentials(workerEndpoint string) (*pb.WorkerCredentials, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/%s", workerEndpoint, "attestation"))
	if err != nil {
		return nil, err
	}

	workerCredentialsBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	workerCredentials := &pb.WorkerCredentials{}
	err = proto.Unmarshal(workerCredentialsBytes, workerCredentials)
	if err != nil {
		return nil, err
	}

	evidence, err := toEvidence(workerCredentials.Attestation)
	if err != nil {
		return nil, err
	}

	workerCredentials.Evidence = evidence

	return workerCredentials, nil
}

func ExecuteEvaluationPack(workerEndpoint string, encryptedEvaluationPack *pb.EncryptedEvaluationPack) ([]byte, error) {
	encryptedEvaluationPackBytes, err := proto.Marshal(encryptedEvaluationPack)
	if err != nil {
		return nil, err
	}

	fmt.Println("Send evaluation pack to worker!")
	resp, err := http.Post(fmt.Sprintf("http://%s/%s", workerEndpoint, "execute-evaluationpack"), "", bytes.NewBuffer(encryptedEvaluationPackBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}
