/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/principal-investigator/pi"
	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/protos"

	"google.golang.org/protobuf/proto"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	startServer()
}

var flagPort string

func init() {
	flag.StringVar(&flagPort, "port", "3002", "Port to listen on")
}

func startServer() {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "x-user")

	r := gin.Default()
	r.Use(cors.New(config))

	// controller
	r.POST("/api/register-study", registerStudy)
	r.POST("/api/approve-experiment", approveExperiment)

	r.Run(":" + flagPort)
}

// RegisterStudyRequest binding from JSON
type RegisterStudyRequest struct {
	StudyId  string   `json:"studyId" binding:"required"`
	Users    []string `json:"users" binding:"required"`
	Metadata string   `json:"metadata"`
}

func registerStudy(c *gin.Context) {
	var req RegisterStudyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userIdentities []*pb.Identity
	for _, user := range req.Users {
		userIdentities = append(userIdentities, pi.CreateIdentity([]byte(user), nil, nil))
	}

	fmt.Printf("Registering study...")
	err := pi.RegisterStudy(req.StudyId, req.Metadata, userIdentities)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// return answer
	c.IndentedJSON(http.StatusOK, "registered")
}

// ApproveExperimentRequest binding from JSON
type ApproveExperimentRequest struct {
	ExperimentId string `json:"experimentId" binding:"required"`
}

func approveExperiment(c *gin.Context) {
	var req ApproveExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Approving experiment...")
	experimentiProto, err := pi.GetExperimentProposal(req.ExperimentId)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	experimentBytes, err := proto.Marshal(experimentiProto)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = pi.DecideOnExperiment(req.ExperimentId, experimentBytes, "approved")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// return answer
	c.IndentedJSON(http.StatusOK, "approved")
}
