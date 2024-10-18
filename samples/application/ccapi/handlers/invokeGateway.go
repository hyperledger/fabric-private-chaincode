package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/ccapi/chaincode"
	"github.com/hyperledger-labs/ccapi/common"
	"github.com/pkg/errors"
)

func InvokeGatewayDefault(c *gin.Context) {
	channelName := os.Getenv("CHANNEL")
	chaincodeName := os.Getenv("CCNAME")

	invokeGateway(c, channelName, chaincodeName)
}

func InvokeGatewayCustom(c *gin.Context) {
	channelName := c.Param("channelName")
	chaincodeName := c.Param("chaincodeName")

	invokeGateway(c, channelName, chaincodeName)
}

func invokeGateway(c *gin.Context, channelName, chaincodeName string) {
	// Get request body
	req := make(map[string]interface{})
	err := c.BindJSON(&req)
	if err != nil {
		common.Abort(c, http.StatusBadRequest, err)
		return
	}

	txName := c.Param("txname")

	// Get endorsers names
	var endorsers []string
	endorsersQuery := c.Query("@endorsers")
	if endorsersQuery != "" {
		endorsersByte, err := base64.StdEncoding.DecodeString(endorsersQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "the @endorsers query parameter must be a base64-encoded JSON array of strings",
			})
			return
		}

		err = json.Unmarshal(endorsersByte, &endorsers)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "the @endorsers query parameter must be a base64-encoded JSON array of strings",
			})
			return
		}
	}

	// Make transient request
	transientMap := make(map[string]interface{})
	for key, value := range req {
		if key[0] == '~' {
			keyTrimmed := strings.TrimPrefix(key, "~")
			transientMap[keyTrimmed] = value
			delete(req, key)
		}
	}

	transientBytes, _ := json.Marshal(transientMap)
	if len(transientMap) == 0 {
		transientMap = nil
	}

	// Make args
	reqBytes, err := json.Marshal(req)
	if err != nil {
		common.Abort(c, http.StatusInternalServerError, errors.Wrap(err, "failed to marshal req body"))
		return
	}

	// Invoke
	user := c.GetHeader("User")
	if user == "" {
		user = "Admin"
	}

	result, err := chaincode.InvokeGateway(channelName, chaincodeName, txName, user, []string{string(reqBytes)}, transientBytes, endorsers)
	if err != nil {
		err, status := common.ParseError(err)
		common.Abort(c, status, err)
		return
	}

	// Parse response
	var payload interface{}
	err = json.Unmarshal(result, &payload)
	if err != nil {
		common.Abort(c, http.StatusInternalServerError, err)
		return
	}

	common.Respond(c, payload, http.StatusOK, nil)
}
