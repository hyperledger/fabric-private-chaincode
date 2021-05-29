/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/experimenter/client/eas"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/experimenter/client/worker"
	"google.golang.org/protobuf/proto"
)

func main() {
	startServer()
}

var flagPort string

func init() {
	flag.StringVar(&flagPort, "port", "3001", "Port to listen on")
}

func startServer() {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "x-user")

	r := gin.Default()
	r.Use(cors.New(config))

	// controller
	r.POST("/api/new-experiment", newExperiment)
	r.POST("/api/execute", executeExperiment)
	r.POST("/api/launch", launchWorker)

	r.Run(":" + flagPort)
}

// NewExperimentRequest binding from JSON
type NewExperimentRequest struct {
	StudyId      string `json:"studyId" binding:"required"`
	ExperimentId string `json:"experimentId" binding:"required"`
}

func newExperiment(c *gin.Context) {
	var req NewExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("register new experiment: %v\n", req)

	workerCredentials, err := worker.GetWorkerCredentials()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = eas.NewExperiment(req.StudyId, req.ExperimentId, workerCredentials)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// return answer
	c.IndentedJSON(http.StatusOK, "created")
}

// ExecuteRequest binding from JSON
type ExecuteRequest struct {
	ExperimentId string `json:"experimentId" binding:"required"`
}

func executeExperiment(c *gin.Context) {
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedEvaluationPack, err := eas.RequestEvaluationPack(req.ExperimentId)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedEvaluationPackBytes, err := proto.Marshal(encryptedEvaluationPack)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resultBytes, err := worker.ExecuteEvaluationPack(encryptedEvaluationPackBytes)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Result received:\n%s", string(resultBytes))

	// return answer
	c.IndentedJSON(http.StatusOK, string(resultBytes))
}

// LaunchResponse binding from JSON
type LaunchResponse struct {
	PublicKey   string `json:"publicKey" binding:"required"`
	Attestation []byte `json:"attestation" binding:"required"`
}

func launchWorker(c *gin.Context) {
	workerCredentials, err := worker.GetWorkerCredentials()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := &LaunchResponse{
		PublicKey:   string(workerCredentials.GetIdentityBytes()),
		Attestation: workerCredentials.GetAttestation(),
	}

	// return answer
	c.IndentedJSON(http.StatusOK, resp)
}
