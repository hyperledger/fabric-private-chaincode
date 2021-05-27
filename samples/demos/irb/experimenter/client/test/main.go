/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/experimenter/client/eas"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/experimenter/client/worker"
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

		encryptedEvaluationPackBytes, err := proto.Marshal(encryptedEvaluationPack)
		if err != nil {
			panic(err)
		}

		resultBytes, err := worker.ExecuteEvaluationPack(encryptedEvaluationPackBytes)
		if err != nil {
			panic(errors.New(err.Error() + ": " + string(resultBytes)))
		}

		fmt.Printf("Result received:\n%s", string(resultBytes))
	}
}
