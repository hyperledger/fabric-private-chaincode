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

var CreateEscrow = transactions.Transaction{
	Tag:         "createEscrow",
	Label:       "Escrow Creation",
	Description: "Creates a new escrow",
	Method:      "POST",
	Callers: []accesscontrol.Caller{
		{
			MSP: "org1MSP",
			OU:  "admin",
		},
		{
			MSP: "org2MSP",
			OU:  "admin",
		},
	},

	Args: []transactions.Argument{
		{
			Tag:         "escrowId",
			Label:       "Escrow ID",
			Description: "ID of Escrow",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:      "buyerPubKey",
			Label:    "Buyer Public Key",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "sellerPubKey",
			Label:    "Seller Public Key",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "amount",
			Label:    "Escrowed Amount",
			DataType: "number",
			Required: true,
		},
		{
			Tag:      "assetType",
			Label:    "Asset Type Reference",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "conditionValue",
			Label:    "Condition Value",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "status",
			Label:    "Escrow Status",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "createdAt",
			Label:    "Creation Timestamp",
			DataType: "datetime",
			Required: false,
		},
		{
			Tag:      "buyerCertHash",
			Label:    "Buyer Certificate Hash",
			DataType: "string",
			Required: true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		escrowId, _ := req["escrowId"].(string)
		buyerPubKey, _ := req["buyerPubKey"].(string)
		sellerPubKey, _ := req["sellerPubKey"].(string)
		amount, _ := req["amount"].(float64)
		assetType, _ := req["assetType"].(string)
		conditionValue, _ := req["conditionValue"].(string)
		status, _ := req["status"].(string)
		buyerCertHash, _ := req["buyerCertHash"].(string)

		escrowMap := make(map[string]interface{})
		escrowMap["@assetType"] = "escrow"
		escrowMap["escrowId"] = escrowId
		escrowMap["buyerPubKey"] = buyerPubKey
		escrowMap["sellerPubKey"] = sellerPubKey
		escrowMap["amount"] = amount
		escrowMap["assetType"] = assetType
		escrowMap["conditionValue"] = conditionValue
		escrowMap["status"] = status
		escrowMap["createdAt"] = time.Now()
		escrowMap["buyerCertHash"] = buyerCertHash

		escrowAsset, err := assets.NewAsset(escrowMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to create escrow asset")
		}

		_, err = escrowAsset.PutNew(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving escrow on blockchain", err.Status())
		}

		assetJSON, nerr := json.Marshal(escrowAsset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode escrow to JSON format")
		}

		return assetJSON, nil
	},
}

var ReadEscrow = transactions.Transaction{
	Tag:         "readEscrow",
	Label:       "Read Escrow",
	Description: "Read an Escrow by its escrowId",
	Method:      "GET",
	Callers: []accesscontrol.Caller{
		{
			MSP: "org1MSP",
			OU:  "admin",
		},
		{
			MSP: "org2MSP",
			OU:  "admin",
		},
	},

	Args: []transactions.Argument{
		{
			Tag:         "escrowId",
			Label:       "Escrow ID",
			Description: "ID of the Escrow to read",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		escrowId, _ := req["escrowId"].(string)

		key := assets.Key{
			"@assetType": "escrow",
			"escrowId":   escrowId,
		}

		asset, err := key.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading escrow from blockchain", err.Status())
		}

		assetJSON, nerr := json.Marshal(asset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode escrow to JSON format")
		}

		return assetJSON, nil
	},
}
