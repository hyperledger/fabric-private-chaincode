package transactions

import (
	"encoding/json"
	"fmt"
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
			Tag:      "balances",
			Label:    "Different Token Balance",
			DataType: "[]number",
			Required: true,
		},
		{
			Tag:      "digitalAssetTypes",
			Label:    "Digital Assets in Holding",
			DataType: "[]->digitalAsset",
			Required: true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		walletId, _ := req["walletId"].(string)
		ownerId, _ := req["ownerId"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)
		balances, _ := req["balances"].([]interface{})
		assetTypes, _ := req["digitalAssetTypes"].([]interface{})

		fmt.Printf("DEBUG: Received assetTypes: %+v\n", assetTypes)
		fmt.Printf("DEBUG: Type of first element: %T\n", assetTypes[0])

		var processedAssetTypes []interface{}
		for _, assetType := range assetTypes {
			switch v := assetType.(type) {
			case string:
				// If it's a string (UUID), convert to proper reference format
				processedAssetTypes = append(processedAssetTypes, map[string]interface{}{
					"@key": "digitalAsset:" + v,
				})
			case map[string]interface{}:
				// If it's already a map (proper reference), use as-is
				processedAssetTypes = append(processedAssetTypes, v)
			default:
				processedAssetTypes = append(processedAssetTypes, v)
			}
		}

		walletMap := make(map[string]interface{})
		walletMap["@assetType"] = "wallet"
		walletMap["walletId"] = walletId
		walletMap["ownerId"] = ownerId
		walletMap["ownerCertHash"] = ownerCertHash
		walletMap["balances"] = balances
		walletMap["digitalAssetTypes"] = processedAssetTypes
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

var GetBalance = transactions.Transaction{
	Tag:         "getBalance",
	Label:       "Get Wallet Balance",
	Description: "Get balance of a specific token in wallet with authentication",
	Method:      "GET",
	// Do we need this?
	Callers: []accesscontrol.Caller{
		{
			MSP: "Org1MSP",
			OU:  "admin",
		},
		{
			MSP: "Org2MSP",
			OU:  "admin",
		},
	},

	Args: []transactions.Argument{
		{
			Tag:         "walletId",
			Label:       "Wallet ID",
			Description: "ID of the wallet",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "assetSymbol",
			Label:       "Asset Symbol",
			Description: "Symbol of the digital asset to check balance for",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "ownerCertHash",
			Label:       "Owner Certificate Hash",
			Description: "Certificate hash for ownership verification",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		walletId, _ := req["walletId"].(string)
		assetSymbol, _ := req["assetSymbol"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)

		// Get wallet
		key := assets.Key{
			"@key": "wallet:" + walletId,
		}

		walletAsset, err := key.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading wallet from blockchain", err.Status())
		}

		// Verify ownership
		if walletAsset.GetProp("ownerCertHash").(string) != ownerCertHash {
			return nil, errors.NewCCError("Unauthorized: Certificate hash mismatch", 403)
		}

		// Find asset index
		digitalAssetTypes := walletAsset.GetProp("digitalAssetTypes").([]interface{})
		balances := walletAsset.GetProp("balances").([]interface{})

		for i, assetRef := range digitalAssetTypes {
			// Get the referenced asset
			var assetKey string
			switch ref := assetRef.(type) {
			case map[string]interface{}:
				assetKey = ref["@key"].(string)
			case string:
				assetKey = "digitalAsset:" + ref
			}

			// Read the asset to get its symbol
			refKey := assets.Key{"@key": assetKey}
			asset, assetErr := refKey.Get(stub)
			if assetErr != nil {
				continue
			}

			if asset.GetProp("symbol").(string) == assetSymbol {
				balance := balances[i].(float64)
				response := map[string]interface{}{
					"walletId":    walletId,
					"assetSymbol": assetSymbol,
					"balance":     balance,
				}
				responseJSON, jsonErr := json.Marshal(response)
				if jsonErr != nil {
					return nil, errors.WrapError(nil, "failed to encode response to JSON format")
				}
				return responseJSON, nil
			}
		}

		return nil, errors.NewCCError("Asset not found in wallet", 404)
	},
}

var GetWalletByOwner = transactions.Transaction{
	Tag:         "getWalletByOwner",
	Label:       "Get Wallet By Owner",
	Description: "Find wallet by owner identity",
	Method:      "GET",
	Callers: []accesscontrol.Caller{
		{
			MSP: "Org1MSP",
			OU:  "admin",
		},
		{
			MSP: "Org2MSP",
			OU:  "admin",
		},
	},

	Args: []transactions.Argument{
		{
			Tag:         "ownerId",
			Label:       "Owner Identity",
			Description: "Identity of the wallet owner",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "ownerCertHash",
			Label:       "Owner Certificate Hash",
			Description: "Certificate hash for authentication",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		ownerId, _ := req["ownerId"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)

		// Search for wallet by owner
		query := map[string]interface{}{
			"selector": map[string]interface{}{
				"@assetType":    "wallet",
				"ownerId":       ownerId,
				"ownerCertHash": ownerCertHash,
			},
		}

		searchResponse, err := assets.Search(stub, query, "", true)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error searching for wallet", 500)
		}

		if len(searchResponse.Result) == 0 {
			return nil, errors.NewCCError("Wallet not found for owner", 404)
		}

		responseJSON, jsonErr := json.Marshal(searchResponse)
		if jsonErr != nil {
			return nil, errors.WrapErrorWithStatus(nil, "Error marshaling response", 500)
		}

		return responseJSON, nil
	},
}

// Need to impose restriction
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

/*
1. Use hash(txnId) => UUID
2. get it from `stub`

---

Split txnID pool among user
	-> Modulo ops

-> Better random num gen
	-> sha256(<Abhinav>_<i=RNG>_<j=counter>)

---

use txnID as UUID of the assets
	-> For multiple asset generated in 1 txn => Deal with it using counter or something
*/
