/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

TODO (eventually):
  eventually refactor the mock backend as it has the potential of wider
  usefulness (the same applies also to the fabric gateway).
  Auction-specific aspects;
   - the bridge change-code has auction in names (trivial to remove)
   - the "/api/getRegisteredUsers" and, in particular,
     "/api/clock_auction/getDefaultAuction", are auction-specific
   - processing of response

  PS: probably also worth moving the calls to __init & __setup as well
  as the unpacking of the payload objects, which are specific to FPC
  to chaincode/fpc_chaincode.go (or handle these calls for non-fpc
  in chaincode/go_chaincode.go such that actual go chaincode doesn't
   have to know about it?)
*/

package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/fabric-private-chaincode/demo/client/backend/mock/api"
	"github.com/hyperledger-labs/fabric-private-chaincode/demo/client/backend/mock/chaincode"
	"github.com/hyperledger-labs/fabric-private-chaincode/utils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

var flagPort string
var flagDebug bool
var stub *MockStubWrapper
var logger = shim.NewLogger("server")
var notifier = NewNotifier()

const ccName = "FPCAuction"
const channelName = "Mychannel"

const mspId = "Org1MSP"
const org = "org1"

func init() {
	flag.StringVar(&flagPort, "port", "3000", "Port to listen on")
	flag.BoolVar(&flagDebug, "debug", false, "debug output")
}

func main() {
	flag.Parse()

	if flagDebug {
		logger.SetLevel(shim.LogDebug)
	}

	// deploy
	deployChaincode()

	// start web service
	startServer()
}

func startServer() {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "x-user")

	r := gin.Default()
	r.Use(cors.New(config))

	// notifications
	r.GET("/api/notifications", notifications)

	// controller
	r.GET("/api/demo/start", startDemo)

	// ledger debug API
	r.GET("/api/ledger", getLedger)
	r.GET("/api/state", getState)
	r.DELETE("/api/state/:key", deleteStateEntry)
	r.POST("/api/state/:key", updateStateEntry)

	// auction util API
	r.GET("/api/getRegisteredUsers", getAllUsers)
	r.GET("/api/clock_auction/getDefaultAuction", getDefaultAuction)
	r.GET("/api/clock_auction/getAuctionDetails/:auctionId", getAuctionDetails)

	// chaincode API
	r.POST("/api/cc/invoke", invoke)
	// note that using a MockStub there is no need to differentiate between query and invoke
	r.POST("/api/cc/query", query)

	r.Run(":" + flagPort)
}

func deployChaincode() {
	logger.Info("Deploy new chaincode")

	stub = NewWrapper(ccName, chaincode.NewMockAuction(), notifier)

	// setup and init
	stub.Creator = "Auctioneer1"
	stub.MockInvoke("someTxID", [][]byte{[]byte("__setup"), []byte("ercc"), []byte(channelName), []byte("tlcc")})
	stub.MockInvoke("1", [][]byte{[]byte("__init"), []byte(ccName)})
}

func notifications(c *gin.Context) {
	listener := notifier.OpenListener()
	defer func() {
		notifier.CloseListener(listener)
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg := <-listener:
			c.SSEvent("update", msg)
		}
		return true
	})
}

func startDemo(c *gin.Context) {
	// destroy enclave
	Destroy(stub)
	notifier.Submit("restart")

	// let's create a new chaincode
	deployChaincode()
	c.IndentedJSON(http.StatusOK, "start")
}

func getLedger(c *gin.Context) {
	ledger := stub.Transactions
	c.IndentedJSON(http.StatusOK, ledger)
}

func getState(c *gin.Context) {
	ledgerState := stub.MockStub.State
	c.IndentedJSON(http.StatusOK, ledgerState)
}

func deleteStateEntry(c *gin.Context) {
	key := c.Params.ByName("key")
	bk, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		panic(err)
	}
	key = string(bk)

	_ = stub.DelState(key)
	c.String(http.StatusOK, "deleted")
}

func updateStateEntry(c *gin.Context) {
	key := c.Params.ByName("key")
	bk, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		panic(err)
	}
	key = string(bk)

	value, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, "error reading data")
	}

	stub.MockStub.TxID = "dummyTXId"
	defer func() { stub.MockStub.TxID = "" }()
	_ = stub.PutState(key, value)

	logger.Infof("updated %s to %s", key, value)
	c.String(http.StatusOK, "updated")
}

func getAllUsers(c *gin.Context) {
	users := api.MockData["getRegisteredUsers"]
	c.IndentedJSON(http.StatusOK, users)
}

func getDefaultAuction(c *gin.Context) {
	auction := api.MockData["getDefaultAuction"]
	c.IndentedJSON(http.StatusOK, auction)
}

func getAuctionDetails(c *gin.Context) {
	auctionId := c.Params.ByName("auctionId")

	val, _ := stub.MockStub.GetState(auctionId)
	if val == nil {
		// no auction created yet
		status := ResponseStatus{
			RC:      1,
			Message: "Does not exist",
		}
		c.IndentedJSON(http.StatusOK, ResponseObject{Status: status})
		return
	}

	resp := ResponseObject{
		Status: ResponseStatus{
			RC:      0,
			Message: "Ok",
		},
		Response: api.MockData["getDefaultAuction"],
	}
	c.IndentedJSON(http.StatusOK, resp)
}

type Payload struct {
	Tx   string
	Args []string
}

// the JSON objects returned from FPC

type ResponseStatus struct {
	RC      int    `json:"rc"`
	Message string `json:"message"`
}

type ResponseObject struct {
	Status   ResponseStatus `json:"status"`
	Response interface{}    `json:"response"`
}

// Unmarshallers for above to ensure the fields exists ...
// (would be nice if there would be a tag 'jsoon:required,...' or alike but alas
// despite 4 years of requests and discussion nothing such has materialized

func (status *ResponseStatus) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		RC      *int    `json:"rc"`
		Message *string `json:"message"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	} else if required.RC == nil || required.Message == nil {
		err = fmt.Errorf("Required fields for ResponseStatus missing")
	} else {
		status.RC = *required.RC
		status.Message = *required.Message
	}
	return
}

func (response *ResponseObject) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		Status *ResponseStatus `json:"status"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	} else if required.Status == nil {
		err = fmt.Errorf("Required fields for ResponseStatus missing")
	} else {
		response.Status = *required.Status
	}
	return
}

// Main invocation handling
func invoke(c *gin.Context) {

	if stub == nil {
		panic("stub is nil!")
	}

	stub.MockStub.ChannelID = channelName
	user := c.GetHeader("x-user")
	stub.Creator = user

	args, err := parsePayload(c)
	if err != nil {
		logger.Error(fmt.Sprintf("Request Error: %s\n", err))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Debugf("%s invokes %s", user, args)
	res := stub.MockInvoke("someTxID", args)

	fpcResponse := createFPCResponse(res)
	c.Data(http.StatusOK, c.ContentType(), fpcResponse)
}

// Main invocation handling
func query(c *gin.Context) {

	if stub == nil {
		panic("stub is nil!")
	}

	stub.MockStub.ChannelID = channelName
	user := c.GetHeader("x-user")
	stub.Creator = user

	args, err := parsePayload(c)
	if err != nil {
		logger.Error(fmt.Sprintf("Request Error: %s\n", err))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Debugf("%s queries %s", user, args)
	res := stub.MockQuery("someTxID", args)

	fpcResponse := createFPCResponse(res)
	c.Data(http.StatusOK, c.ContentType(), fpcResponse)
}

func createFPCResponse(res peer.Response) []byte {

	// NOTE: we (try to) return error even if the invocation get success back
	// but does not contain a response payload. According to the auction
	// specifications, all queries and transactions should return a response
	// object (even more specifically, an object which at the very least
	// contains a 'status' field)
	var fpcResponse []byte
	var errMsg *string = nil // nil means no error
	// we might get payload and response regardless of invocation success,
	// so try to decode in all cases
	if res.Payload != nil {
		var response utils.Response
		// unwarp ecc response and return only responseData
		if err := json.Unmarshal(res.Payload, &response); err != nil {
			msg := fmt.Sprintf("No valid response payload received due to error=%v (status=%v/message=%v)",
				err, res.Status, res.Message)
			errMsg = &msg
		} else {
			logger.Debugf("FPC response: ResponseData='%s'",
				response.ResponseData)
			fpcResponse = response.ResponseData
			// a proper client would now also verify response signature,
			// we just make sure the response is a json object as expected
			var responseObj ResponseObject
			if err = json.Unmarshal(fpcResponse, &responseObj); err != nil {
				msg := fmt.Sprintf("Response payload '%s' not a valid response object (status=%v/message=%v)",
					fpcResponse, res.Status, res.Message)
				errMsg = &msg
			}
		}
	} else {
		msg := fmt.Sprintf("No response payload received (status=%v/message=%v)",
			res.Status, res.Message)
		errMsg = &msg
	}

	if errMsg != nil {
		fpcResponseJson := ResponseObject{
			Status: ResponseStatus{
				RC:      499, // TODO (maybe): more specific explicit error codes?
				Message: *errMsg,
			},
			Response: fpcResponse,
		}
		fpcResponse, _ = json.Marshal(fpcResponseJson)
	}
	return fpcResponse
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
