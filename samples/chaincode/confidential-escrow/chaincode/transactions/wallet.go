package transactions

import (
	"encoding/json"
	"time"

	"github.com/hyperledger-labs/cc-tools/accesscontrol"
	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/errors"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	"github.com/hyperledger-labs/cc-tools/transactions"
)

var CreateWallet = transactions.Transaction{
	Tag:         "createWallet",
	Label:       "Wallet Creation",
	Description: "Creates a new Wallet",
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
			Tag:         "walletId",
			Label:       "Wallet ID",
			Description: "ID of Wallet",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:      "ownerId",
			Label:    "Owner Identity",
			DataType: "string",
			Required: true,
		},
		{
			Tag:         "ownerCertHash",
			Label:       "Owner Certificate Hash",
			Description: "Hash of Owner's Certificate who created this wallet",
			DataType:    "string",
			Required:    true, // testing purpose
		},
		{
			Tag:      "balance",
			Label:    "Token Balance",
			DataType: "number",
			Required: true,
		},
		{
			Tag:      "digitalAssetType",
			Label:    "Digital Asset in Holding",
			DataType: "string",
			Required: true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		// Add this debug line at the start
		// creator, _ := stub.Get
		// fmt.Printf("DEBUG: Client identity: %s\n", string(creator))

		walletId, _ := req["walletId"].(string)
		ownerId, _ := req["ownerId"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)
		balance, _ := req["balance"].(float64)
		assetType, _ := req["digitalAssetType"].(string)

		walletMap := make(map[string]interface{})
		walletMap["@assetType"] = "wallet"
		walletMap["walletId"] = walletId
		walletMap["ownerId"] = ownerId
		walletMap["ownerCertHash"] = ownerCertHash
		walletMap["balance"] = balance
		walletMap["digitalAssetType"] = assetType
		walletMap["createdAt"] = time.Now()

		walletAsset, err := assets.NewAsset(walletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to create wallet asset")
		}

		_, err = walletAsset.PutNew(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving wallet on blockchain", err.Status())
		}

		assetJSON, nerr := json.Marshal(walletAsset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode wallet to JSON format")
		}

		return assetJSON, nil
	},
}

var ReadWallet = transactions.Transaction{
	Tag:         "readWallet",
	Label:       "Read Wallet",
	Description: "Read a Wallet by its walletId",
	Method:      "GET",
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
			Tag:         "uuid",
			Label:       "UUID",
			Description: "UUID of the Digital Asset to read",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		uuid, _ := req["uuid"].(string)

		key := assets.Key{
			"@key": "wallet:" + uuid,
		}

		asset, err := key.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading wallet from blockchain", err.Status())
		}

		assetJSON, nerr := json.Marshal(asset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode wallet to JSON format")
		}

		return assetJSON, nil
	},
}
