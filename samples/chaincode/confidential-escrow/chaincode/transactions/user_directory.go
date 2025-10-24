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
			MSP: "Org1MSP",
			OU:  "admin",
		}, {
			MSP: "Org2MSP",
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
			Tag:      "walletUUID",
			Label:    "Associated Wallet UUID",
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
		walletId, _ := req["walletUUID"].(string)
		certHash, _ := req["certHash"].(string)

		userDirMap := make(map[string]interface{})
		userDirMap["@assetType"] = "userdir"
		userDirMap["publicKeyHash"] = publicKeyHash
		userDirMap["walletUUID"] = walletId
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
		{MSP: "Org1MSP", OU: "admin"},
		{MSP: "Org2MSP", OU: "admin"},
	},

	Args: []transactions.Argument{
		{
			Tag:         "userDir",
			Label:       "User Directory",
			Description: "User Directory to read",
			DataType:    "->userdir",
			Required:    true,
		},
		{
			Tag:         "certHash",
			Label:       "Certificate Hash",
			Description: "Certificate hash for ownership verification",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		userDirRef, _ := req["userDir"].(assets.Key)
		certHash, _ := req["certHash"].(string)

		asset, err := userDirRef.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading user directory entry from blockchain", err.Status())
		}

		// Verify ownership
		storedCertHash := asset.GetProp("certHash").(string)
		if storedCertHash != certHash {
			return nil, errors.NewCCError("Unauthorized: Certificate hash mismatch", 403)
		}

		assetJSON, nerr := json.Marshal(asset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		return assetJSON, nil
	},
}

// var ReadUserDir = transactions.Transaction{
// 	Tag:         "readUserDir",
// 	Label:       "Read User Directory",
// 	Description: "Read a User Directory by its publicKeyHash with authentication",
// 	Method:      "GET",
// 	Callers: []accesscontrol.Caller{
// 		{
// 			MSP: "Org1MSP",
// 			OU:  "admin",
// 		}, {
// 			MSP: "Org2MSP",
// 			OU:  "admin",
// 		},
// 	},
//
// 	Args: []transactions.Argument{
// 		{
// 			Tag:         "uuid",
// 			Label:       "UUID",
// 			Description: "UUID of the User Directory to read",
// 			DataType:    "string",
// 			Required:    true,
// 		},
// 		{
// 			Tag:         "certHash",
// 			Label:       "Certificate Hash",
// 			Description: "Certificate hash for ownership verification",
// 			DataType:    "string",
// 			Required:    true,
// 		},
// 	},
//
// 	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
// 		uuid, _ := req["uuid"].(string)
// 		certHash, _ := req["certHash"].(string)
//
// 		key := assets.Key{
// 			"@key": "userdir:" + uuid,
// 		}
//
// 		asset, err := key.Get(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error reading user directory entry from blockchain", err.Status())
// 		}
//
// 		// Verify ownership - only the owner can read their own directory
// 		storedCertHash := asset.GetProp("certHash").(string)
// 		if storedCertHash != certHash {
// 			return nil, errors.NewCCError("Unauthorized: Certificate hash mismatch", 403)
// 		}
//
// 		assetJSON, nerr := json.Marshal(asset)
// 		if nerr != nil {
// 			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
// 		}
//
// 		return assetJSON, nil
// 	},
// }
