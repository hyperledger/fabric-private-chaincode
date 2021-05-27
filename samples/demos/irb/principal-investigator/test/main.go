/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/principal-investigator/pi"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/protos"
)

func RegisterStudy() error {
	users := []string{"user1", "user2", "user3", "user4", "user5", "user6"}
	userIdentities := []*pb.Identity{}

	for i := 0; i < len(users); i++ {
		userIdentities = append(userIdentities, pi.CreateIdentity([]byte(users[i]), nil, nil))
	}

	fmt.Printf("Registering study...")
	err := pi.RegisterStudy("study1", "", userIdentities)
	if err != nil {
		return err
	}

	fmt.Printf("done.\n")
	return nil
}

func ApproveExperiment() error {
	fmt.Printf("Approving experiment...")
	experimentiProto, err := pi.GetExperimentProposal("experiment1")
	if err != nil {
		return err
	}

	experimentBytes, err := proto.Marshal(experimentiProto)
	if err != nil {
		return err
	}

	err = pi.DecideOnExperiment("experiment1", experimentBytes, "approved")
	if err != nil {
		return err
	}

	fmt.Printf("done.\n")
	return nil
}

func usage() {
	fmt.Printf("Usage: use one parameter \"registerstudy\" or \"approveexperiment\"")
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(-1)
	}

	if os.Args[1] == "registerstudy" {
		err := RegisterStudy()
		if err != nil {
			panic(err)
		}
	} else if os.Args[1] == "approveexperiment" {
		err := ApproveExperiment()
		if err != nil {
			panic(err)
		}
	} else {
		usage()
		os.Exit(-2)
	}
}
