package transactions

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger-labs/cc-tools/accesscontrol"
	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/errors"
	"github.com/hyperledger-labs/cc-tools/events"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	"github.com/hyperledger-labs/cc-tools/transactions"
)

var CreateUserDir = transactions.Transaction{
	Tag:         "createUserDir",
	Label:       "User Directory Creation",
	Description: "Creates a new User entry",
	Method:      "POST",
	Callers: []accesscontrol.Caller{
		{
			MSP: "org1MSP",
			OU:  "admin",
		}, {
			MSP: "org2MSP",
			OU:  "admin",
		},
	},

	Args: []transactions.Argument{
		{
			Tag:      "publicKeyHash",
			Label:    "Public Key Hash",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "walletId",
			Label:    "Associated Wallet ID",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "certHash",
			Label:    "Certificate Hash",
			DataType: "string",
			Required: true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		publicKeyHash, _ := req["publicKeyHash"].(string)
		walletId, _ := req["walletId"].(string)
		certHash, _ := req["certHash"].(string)

		userDirMap := make(map[string]interface{})
		userDirMap["@assetType"] = "userdir"
		userDirMap["publicKeyHash"] = publicKeyHash
		userDirMap["walletId"] = walletId
		userDirMap["certHash"] = certHash

		userDirAsset, err := assets.NewAsset(userDirMap)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading user directory entry from blockchain", err.Status())
		}

		_, err = userDirAsset.PutNew(stub)
		if err != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		assetJson, nerr := json.Marshal(userDirAsset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		logMsg, ok := json.Marshal(fmt.Sprintf("New  user directory created: %s", publicKeyHash))
		if ok != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		events.CallEvent(stub, "createUserDirLog", logMsg)

		return assetJson, nil
	},
}

var ReadUserDir = transactions.Transaction{
	Tag:         "readUserDir",
	Label:       "Read User Directory",
	Description: "Read a User Directory by its publicKeyHash",
	Method:      "GET",
	Callers: []accesscontrol.Caller{
		{
			MSP: "org1MSP",
			OU:  "admin",
		}, {
			MSP: "org2MSP",
			OU:  "admin",
		},
	},

	Args: []transactions.Argument{
		{
			Tag:         "publicKeyHash",
			Label:       "Public Key Hash",
			Description: "Hash of Public Key to read",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		publicKeyHash, _ := req["publicKeyHash"].(string)

		key := assets.Key{
			"@assetType":    "userdir",
			"publicKeyHash": publicKeyHash,
		}

		asset, err := key.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error user directory entry from blockchain", err.Status())
		}

		assetJSON, nerr := json.Marshal(asset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		return assetJSON, nil
	},
}
