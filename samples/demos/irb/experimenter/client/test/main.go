/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/experimenter/client/eas"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/experimenter/client/worker"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/protos"
)

func usage() {
	fmt.Printf("Usage: use one parameter \"newexperiment\" or \"executeevaluationpack\"")
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(-1)
	}

	if os.Args[1] == "newexperiment" {
		workerCredentials, err := worker.GetWorkerCredentials()
		if err != nil {
			panic(err)
		}

		err = eas.NewExperiment("study1", "experiment1", workerCredentials)
		if err != nil {
			panic(err)
		}
	}

	if os.Args[1] == "executeevaluationpack" {
		encryptedEvaluationPack, err := eas.RequestEvaluationPack("experiment1")
		if err != nil {
			panic(err)
		}

		//no encryption
		evaluationPackBytes := encryptedEvaluationPack.GetEncryptedEvaluationpack()
		evaluationPackMessage := &pb.EvaluationPackMessage{}
		err = proto.Unmarshal(evaluationPackBytes, evaluationPackMessage)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Dumping registered data decryption keys keys:\n")
		registeredData := evaluationPackMessage.GetRegisteredData()
		for i := 0; i < len(registeredData); i++ {
			dk := registeredData[i].GetDecryptionKey()
			fmt.Printf("%d: %s\n", i, string(dk))
		}

		encryptedEvaluationPackBytes, err := proto.Marshal(encryptedEvaluationPack)
		if err != nil {
			panic(err)
		}

		resultBytes, err := worker.ExecuteEvaluationPack(encryptedEvaluationPackBytes)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Result received:\n%s", string(resultBytes))
	}
}
