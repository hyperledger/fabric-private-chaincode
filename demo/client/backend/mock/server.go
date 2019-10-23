/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/fabric-private-chaincode/demo/client/backend/mock/api"
	"github.com/hyperledger-labs/fabric-private-chaincode/demo/client/backend/mock/chaincode"
	"github.com/hyperledger-labs/fabric-private-chaincode/utils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var flagPort string
var stub *shim.MockStub
var logger = shim.NewLogger("server")

const ccName = "ecc"

func init() {
	flag.StringVar(&flagPort, "port", "3000", "Port to listen on")
}

func main() {
	flag.Parse()

	stub = shim.NewMockStub(ccName, chaincode.NewMockAuction())

	// setup and init
	stub.MockInvoke("someTxID", [][]byte{[]byte("__setup"), []byte("ercc"), []byte("mychannel"), []byte("tlcc")})
	stub.MockInvoke("1", [][]byte{[]byte("__init"), []byte("My Auction")})

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "x-user")

	r := gin.Default()
	r.Use(cors.New(config))

	r.GET("/api/getRegisteredUsers", getAllUsers)
	r.GET("/api/clock_auction/getDefaultAuction", getDefaultAuction)
	r.POST("/api/cc/invoke", invoke)
	// note that using a MockStub there is no need to differentiate between query and invoke
	r.POST("/api/cc/query", invoke)

	r.Run(":" + flagPort)
}

func getAllUsers(c *gin.Context) {
	users := api.MockData["getRegisteredUsers"]
	c.IndentedJSON(http.StatusOK, users)
}

func getDefaultAuction(c *gin.Context) {
	users := api.MockData["getDefaultAuction"]
	c.IndentedJSON(http.StatusOK, users)
}

type Payload struct {
	Tx   string
	Args []string
}

func invoke(c *gin.Context) {

	user := c.GetHeader("x-user")
	logger.Debug(fmt.Sprintf("user: %s\n", user))

	// TODO set creator

	args, err := parsePayload(c)
	if err != nil {
		//fmt.Println("Request Error: " + err.Error())
		logger.Error(fmt.Sprintf("Request Error: %s\n", err))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := stub.MockInvoke("someTxID", args)
	if res.Status != shim.OK {
		//fmt.Printf("Chaincode error: %s", res.Message)
		logger.Error(fmt.Sprintf("Chaincode Error: %s\n", res.Message))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": res.Message})
		return
	}

	// unwarp ecc response and return only responseData
	// a proper client would now also verify response signature
	var response utils.Response
	err = json.Unmarshal(res.Payload, &response)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.Data(http.StatusOK, c.ContentType(), response.ResponseData)
}

func parsePayload(c *gin.Context) ([][]byte, error) {
	var payload Payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		return nil, err
	}

	args := make([][]byte, 0)
	args = append(args, []byte(payload.Tx))
	for _, b := range payload.Args {
		args = append(args, []byte(b))
	}

	return args, nil
}
