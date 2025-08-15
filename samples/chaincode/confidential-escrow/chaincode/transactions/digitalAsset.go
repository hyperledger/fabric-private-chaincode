package transactions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger-labs/cc-tools/accesscontrol"
	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/errors"
	"github.com/hyperledger-labs/cc-tools/events"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	"github.com/hyperledger-labs/cc-tools/transactions"
)

var CreateDigitalAsset = transactions.Transaction{
	Tag:         "createDigitalAsset",
	Label:       "Digital Asset Creation",
	Description: "Creates a new Digital Asset e.g. CBDC Tokens",
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
			Tag:         "name",
			Label:       "Name",
			Description: "Name of the Digital Asset",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "symbol",
			Label:       "Symbol",
			Description: "Symbol of the Digital Asset",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "decimals",
			Label:       "Decimal Places",
			Description: "Decimal Places in Digital Asset",
			DataType:    "number",
			Required:    true,
		},
		{
			Tag:         "totalSupply",
			Label:       "Total Supply",
			Description: "Total Supply of the Digital Asset",
			DataType:    "number",
			Required:    true,
		},
		{
			Tag:         "owner",
			Label:       "Owner Identity",
			Description: "Identitiy of Digital Asset's creator",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "issuedAt",
			Label:       "Issued At",
			Description: "Time at which this token was created",
			DataType:    "datetime",
			Required:    false,
		},
		{
			Tag:         "issuerHash",
			Label:       "Issuer Certificate Hash",
			Description: "Hash of Issuer's Certificate who created this Digital Asset",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		name, _ := req["name"].(string)
		symbol, _ := req["symbol"].(string)
		decimals, _ := req["decimals"].(float64)
		totalSupply, _ := req["totalSupply"].(float64)
		owner, _ := req["owner"].(string)
		issuerHash, _ := req["issuerHash"].(string)

		assetMap := make(map[string]interface{})
		assetMap["@assetType"] = "digitalAsset"
		assetMap["name"] = name
		assetMap["symbol"] = symbol
		assetMap["decimals"] = decimals
		assetMap["totalSupply"] = totalSupply
		assetMap["owner"] = owner
		assetMap["issuedAt"] = time.Now()
		assetMap["issuerHash"] = issuerHash

		digitalAsset, err := assets.NewAsset(assetMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to create digital asset")
		}

		_, err = digitalAsset.PutNew(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving digital asset on blockchain", err.Status())
		}

		assetJSON, nerr := json.Marshal(digitalAsset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		logMsg, ok := json.Marshal(fmt.Sprintf("New Digital Asset created: %s", name))
		if ok != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		events.CallEvent(stub, "createDigitalAssetLog", logMsg)

		return assetJSON, nil
	},
}

var ReadDigitalAsset = transactions.Transaction{
	Tag:         "readDigitalAsset",
	Label:       "Read Digital Asset",
	Description: "Read a Digital Asset by its symbol",
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
			Tag:         "symbol",
			Label:       "Symbol",
			Description: "Symbol of the Digital Asset to read",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		symbol, _ := req["symbol"].(string)

		key := assets.Key{
			"@assetType": "digitalAsset",
			"symbol":     symbol,
		}

		asset, err := key.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading digital asset from blockchain", err.Status())
		}

		assetJSON, nerr := json.Marshal(asset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		return assetJSON, nil
	},
}
