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
	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	dp "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/data-provider"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/data-provider/service/pdf"
)

func main() {
	startServer()
}

var flagPort string

func init() {
	flag.StringVar(&flagPort, "port", "3000", "Port to listen on")
}

func startServer() {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "x-user")

	r := gin.Default()
	r.Use(cors.New(config))

	// controller
	r.POST("/api/upload", upload)

	initDummyPatients()

	r.Run(":" + flagPort)
}

// UploadRequest binding from JSON
type UploadRequest struct {
	Data       []byte   `json:"data" binding:"required"`
	DataName   string   `json:"dataName" binding:"required"`
	Domain     string   `json:"domain" binding:"required"`
	AllowedUse []string `json:"allowedUse" binding:"required"`
}

func initDummyPatients() {
	users := []string{"patient1", "patient2"}

	fmt.Printf("Creating patients and data...")
	for i := 0; i < len(users); i++ {
		_, _, _, err := dp.LoadOrCreateUser(users[i])
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("done.\n")
}

func upload(c *gin.Context) {

	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// parse PDF
	res, err := pdf.ParseQuestionForm(req.Data)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// read study id

	// lookup user
	uuid, _, vk, err := dp.LoadUser(res.UUID)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	answerData := res.Answers.ToString()

	// encrypt with new random key
	sk, err := crypto.NewSymmetricKey()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Encrypting data ... ")
	encryptedData, err := crypto.EncryptMessage(sk, []byte(answerData))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("done\n")

	// upload encrypted data
	handle, err := dp.Upload(encryptedData)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// register consent at FPC IRB
	err = dp.RegisterData(res.StudyId, uuid, vk, sk, handle)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// return answer
	c.IndentedJSON(http.StatusOK, "uploaded")
}
