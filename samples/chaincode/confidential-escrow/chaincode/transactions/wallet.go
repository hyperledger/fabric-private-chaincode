package transactions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
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
			Tag:      "ownerPubKey",
			Label:    "Owner Public Key",
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
		// {
		// 	Tag:      "balances",
		// 	Label:    "Different Token Balance",
		// 	DataType: "[]number",
		// 	Required: true,
		// },
		// {
		// 	Tag:      "digitalAssetTypes",
		// 	Label:    "Digital Assets in Holding",
		// 	DataType: "[]->digitalAsset",
		// 	Required: true,
		// },
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		walletId, _ := req["walletId"].(string)
		ownerPublicKey, _ := req["ownerPubKey"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)

		hash := sha256.Sum256([]byte(ownerPublicKey))
		pubKeyHash := hex.EncodeToString(hash[:])
		fmt.Printf("DEBUG: Owner PubKey: %s\n", ownerPublicKey)
		fmt.Printf("DEBUG: Owner PubKey Hash: %s\n", pubKeyHash)
		fmt.Printf("DEBUG: Creating UserDir with key: userdir:%s\n", pubKeyHash)

		// balances, _ := req["balances"].([]interface{})
		// assetTypes, _ := req["digitalAssetTypes"].([]interface{})

		// fmt.Printf("DEBUG: Received assetTypes: %+v\n", assetTypes)
		// fmt.Printf("DEBUG: Type of first element: %T\n", assetTypes[0])
		//
		// var processedAssetTypes []interface{}
		// for _, assetType := range assetTypes {
		// 	switch v := assetType.(type) {
		// 	case string:
		// 		// If it's a string (UUID), convert to proper reference format
		// 		processedAssetTypes = append(processedAssetTypes, map[string]interface{}{
		// 			"@key": "digitalAsset:" + v,
		// 		})
		// 	case map[string]interface{}:
		// 		// If it's already a map (proper reference), use as-is
		// 		processedAssetTypes = append(processedAssetTypes, v)
		// 	default:
		// 		processedAssetTypes = append(processedAssetTypes, v)
		// 	}
		// }

		walletMap := make(map[string]interface{})
		walletMap["@assetType"] = "wallet"
		walletMap["walletId"] = walletId
		walletMap["ownerPubKey"] = ownerPublicKey
		walletMap["ownerCertHash"] = ownerCertHash
		walletMap["escrowBalances"] = make([]interface{}, 0)
		walletMap["balances"] = make([]interface{}, 0)
		walletMap["digitalAssetTypes"] = make([]interface{}, 0)
		walletMap["createdAt"] = time.Now()

		walletAsset, err := assets.NewAsset(walletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to create wallet asset")
		}

		// _, err = walletAsset.PutNew(stub)
		_, err = walletAsset.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving wallet on blockchain", err.Status())
		}

		// Create corresponding UserDir entry
		walletUUID := strings.Split(walletAsset.GetProp("@key").(string), ":")[1]

		userDirMap := make(map[string]interface{})
		userDirMap["@assetType"] = "userdir"
		userDirMap["publicKeyHash"] = pubKeyHash // Using certHash as identifier
		userDirMap["walletUUID"] = walletUUID
		userDirMap["certHash"] = ownerCertHash

		userDirAsset, err := assets.NewAsset(userDirMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to create user directory")
		}

		_, err = userDirAsset.PutNew(stub)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to save user directory")
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
			Tag:         "walletUUID",
			Label:       "Wallet UUID",
			Description: "UUID of the wallet",
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
		walletId, _ := req["walletUUID"].(string)
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

var GetEscrowBalance = transactions.Transaction{
	Tag:         "getEscrowBalance",
	Label:       "Get Wallet Escrow Balance",
	Description: "Get escrowed balance of a specific token in wallet",
	Method:      "GET",
	Callers: []accesscontrol.Caller{
		{MSP: "Org1MSP", OU: "admin"},
		{MSP: "Org2MSP", OU: "admin"},
	},
	Args: []transactions.Argument{
		{Tag: "walletUUID", DataType: "string", Required: true},
		{Tag: "assetSymbol", DataType: "string", Required: true},
		{Tag: "ownerCertHash", DataType: "string", Required: true},
	},
	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		walletId, _ := req["walletUUID"].(string)
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
		escrowBalances := walletAsset.GetProp("escrowBalances").([]interface{})

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
				escrowBalance := escrowBalances[i].(float64)
				response := map[string]interface{}{
					"walletId":      walletId,
					"assetSymbol":   assetSymbol,
					"escrowBalance": escrowBalance,
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
	Description: "Find wallet by providing wallet UUID directly",
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
			Tag:         "walletUUID",
			Label:       "Wallet UUID",
			Description: "UUID of the wallet to find",
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
		walletUuid, _ := req["walletUUID"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)

		// Get wallet directly
		walletKey := assets.Key{"@key": "wallet:" + walletUuid}
		wallet, err := walletKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Wallet not found", 404)
		}

		// Verify ownership
		if wallet.GetProp("ownerCertHash").(string) != ownerCertHash {
			return nil, errors.NewCCError("Unauthorized: Certificate hash mismatch", 403)
		}

		responseJSON, jsonErr := json.Marshal(wallet)
		if jsonErr != nil {
			return nil, errors.WrapError(nil, "failed to encode wallet to JSON format")
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
			Description: "UUID of the wallet to read",
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
