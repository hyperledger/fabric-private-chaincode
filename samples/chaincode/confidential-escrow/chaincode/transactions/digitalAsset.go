package transactions

import (
	"encoding/json"
	"fmt"
	"strings"
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
			MSP: "Org1MSP",
			OU:  "admin",
		}, {
			MSP: "Org2MSP",
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
			MSP: "Org1MSP",
			OU:  "admin",
		}, {
			MSP: "Org2MSP",
			OU:  "admin",
		},
	},

	Args: []transactions.Argument{
		// need to find a better way other than UUID.. i.e. search via Symbol or something
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
			"@key": "digitalAsset:" + uuid,
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

var MintTokens = transactions.Transaction{
	Tag:         "mintTokens",
	Label:       "Mint Tokens",
	Description: "Mint new tokens to a wallet (issuer only)",
	Method:      "POST",
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
			Tag:         "assetId",
			Label:       "Asset ID",
			Description: "ID of the digital asset",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "walletUUID",
			Label:       "Target Wallet UUID",
			Description: "Wallet to mint tokens to",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "amount",
			Label:       "Amount to Mint",
			Description: "Number of tokens to mint",
			DataType:    "number",
			Required:    true,
		},
		{
			Tag:         "issuerCertHash",
			Label:       "Issuer Certificate Hash",
			Description: "Certificate hash for issuer verification",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		assetId, _ := req["assetId"].(string)
		walletUUID, _ := req["walletUUID"].(string)
		amount, _ := req["amount"].(float64)
		issuerCertHash, _ := req["issuerCertHash"].(string)

		// Verify issuer authorization
		assetKey := assets.Key{"@key": "digitalAsset:" + assetId}
		asset, err := assetKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading digital asset", err.Status())
		}

		if asset.GetProp("issuerHash").(string) != issuerCertHash {
			return nil, errors.NewCCError("Unauthorized: Only asset issuer can mint tokens", 403)
		}

		// Get wallet
		walletKey := assets.Key{"@key": "wallet:" + walletUUID}
		walletAsset, err := walletKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading wallet", err.Status())
		}

		digitalAssetTypes := walletAsset.GetProp("digitalAssetTypes").([]interface{})
		balances := walletAsset.GetProp("balances").([]interface{})
		escrowBalances := walletAsset.GetProp("escrowBalances").([]interface{})

		// Find asset index and update balance
		assetFound := false
		for i, assetRef := range digitalAssetTypes {
			var refAssetId string
			switch ref := assetRef.(type) {
			case map[string]interface{}:
				refAssetId = strings.Split(ref["@key"].(string), ":")[1]
			case string:
				refAssetId = ref
			}

			if refAssetId == assetId {
				currentBalance := balances[i].(float64)
				balances[i] = currentBalance + amount
				assetFound = true
				break
			}
		}

		if !assetFound {
			digitalAssetTypes = append(digitalAssetTypes, map[string]interface{}{
				"@key": "digitalAsset:" + assetId,
			})
			balances = append(balances, amount)
			escrowBalances = append(escrowBalances, 0.0)
		}

		// Create updated wallet map
		walletMap := make(map[string]interface{})
		walletMap["@assetType"] = "wallet"
		walletMap["@key"] = "wallet:" + walletUUID
		walletMap["walletId"] = walletAsset.GetProp("walletId")
		walletMap["ownerPubKey"] = walletAsset.GetProp("ownerPubKey")
		walletMap["ownerCertHash"] = walletAsset.GetProp("ownerCertHash")
		walletMap["balances"] = balances
		walletMap["escrowBalances"] = escrowBalances
		walletMap["digitalAssetTypes"] = digitalAssetTypes
		walletMap["createdAt"] = walletAsset.GetProp("createdAt")

		updatedWallet, err := assets.NewAsset(walletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update wallet")
		}

		_, err = updatedWallet.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error updating wallet", err.Status())
		}

		// Update total supply
		currentSupply := asset.GetProp("totalSupply").(float64)
		assetMap := make(map[string]interface{})
		assetMap["@assetType"] = "digitalAsset"
		assetMap["@key"] = "digitalAsset:" + assetId
		assetMap["name"] = asset.GetProp("name")
		assetMap["symbol"] = asset.GetProp("symbol")
		assetMap["decimals"] = asset.GetProp("decimals")
		assetMap["totalSupply"] = currentSupply + amount
		assetMap["owner"] = asset.GetProp("owner")
		assetMap["issuedAt"] = asset.GetProp("issuedAt")
		assetMap["issuerHash"] = asset.GetProp("issuerHash")

		updatedAsset, err := assets.NewAsset(assetMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update asset")
		}

		_, err = updatedAsset.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error updating asset", err.Status())
		}

		response := map[string]interface{}{
			"message":     "Tokens minted successfully",
			"assetId":     assetId,
			"walletId":    walletUUID,
			"amount":      amount,
			"totalSupply": currentSupply + amount,
		}

		respJSON, jsonErr := json.Marshal(response)
		if jsonErr != nil {
			return nil, errors.WrapError(nil, "failed to encode response to JSON format")
		}

		return respJSON, nil
	},
}

var TransferTokens = transactions.Transaction{
	Tag:         "transferTokens",
	Label:       "Transfer Tokens",
	Description: "Transfer tokens between wallets with balance validation",
	Method:      "POST",
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
			Tag:         "fromWalletUUID",
			Label:       "From Wallet UUID",
			Description: "Source wallet UUID",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "toWalletUUID",
			Label:       "To Wallet UUID",
			Description: "Destination wallet UUID",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "assetId",
			Label:       "Asset ID",
			Description: "ID of the digital asset to transfer",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "amount",
			Label:       "Transfer Amount",
			Description: "Number of tokens to transfer",
			DataType:    "number",
			Required:    true,
		},
		{
			Tag:         "senderCertHash",
			Label:       "Sender Certificate Hash",
			Description: "Certificate hash of the sender for authorization",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		fromWalletUUID, _ := req["fromWalletUUID"].(string)
		toWalletUUID, _ := req["toWalletUUID"].(string)
		assetId, _ := req["assetId"].(string)
		amount, _ := req["amount"].(float64)
		senderCertHash, _ := req["senderCertHash"].(string)

		// Get source wallet
		fromKey := assets.Key{"@key": "wallet:" + fromWalletUUID}
		fromWalletAsset, err := fromKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading source wallet", err.Status())
		}

		// Verify sender authorization
		if fromWalletAsset.GetProp("ownerCertHash").(string) != senderCertHash {
			return nil, errors.NewCCError("Unauthorized: Sender certificate mismatch", 403)
		}

		// Get destination wallet
		toKey := assets.Key{"@key": "wallet:" + toWalletUUID}
		toWalletAsset, err := toKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading destination wallet", err.Status())
		}

		// Update source wallet balance
		fromAssetTypes := fromWalletAsset.GetProp("digitalAssetTypes").([]interface{})
		fromBalances := fromWalletAsset.GetProp("balances").([]interface{})
		fromEscrowBalances := fromWalletAsset.GetProp("escrowBalances").([]interface{})

		fromAssetFound := false
		for i, assetRef := range fromAssetTypes {
			var refAssetId string
			switch ref := assetRef.(type) {
			case map[string]interface{}:
				refAssetId = strings.Split(ref["@key"].(string), ":")[1]
			case string:
				refAssetId = ref
			}

			if refAssetId == assetId {
				currentBalance := fromBalances[i].(float64)
				if currentBalance < amount {
					return nil, errors.NewCCError("Insufficient balance", 400)
				}
				fromBalances[i] = currentBalance - amount
				fromAssetFound = true
				break
			}
		}

		if !fromAssetFound {
			return nil, errors.NewCCError("Asset not found in source wallet", 404)
		}

		// Update destination wallet balance
		toAssetTypes := toWalletAsset.GetProp("digitalAssetTypes").([]interface{})
		toBalances := toWalletAsset.GetProp("balances").([]interface{})
		toEscrowBalances := toWalletAsset.GetProp("escrowBalances").([]interface{})

		toAssetFound := false
		for i, assetRef := range toAssetTypes {
			var refAssetId string
			switch ref := assetRef.(type) {
			case map[string]interface{}:
				refAssetId = strings.Split(ref["@key"].(string), ":")[1]
			case string:
				refAssetId = ref
			}

			if refAssetId == assetId {
				currentBalance := toBalances[i].(float64)
				toBalances[i] = currentBalance + amount
				toAssetFound = true
				break
			}
		}

		// PPS: if asset not found, Add asset
		if !toAssetFound {
			toAssetTypes = append(toAssetTypes, map[string]interface{}{
				"@key": "digitalAsset:" + assetId,
			})
			toBalances = append(toBalances, amount)
			toEscrowBalances = append(toEscrowBalances, 0.0)
		}

		// Save updated source wallet
		fromWalletMap := make(map[string]interface{})
		fromWalletMap["@assetType"] = "wallet"
		fromWalletMap["@key"] = "wallet:" + fromWalletUUID
		fromWalletMap["walletId"] = fromWalletAsset.GetProp("walletId")
		fromWalletMap["ownerPubKey"] = fromWalletAsset.GetProp("ownerPubKey")
		fromWalletMap["ownerCertHash"] = fromWalletAsset.GetProp("ownerCertHash")
		fromWalletMap["balances"] = fromBalances
		fromWalletMap["escrowBalances"] = fromEscrowBalances
		fromWalletMap["digitalAssetTypes"] = fromAssetTypes
		fromWalletMap["createdAt"] = fromWalletAsset.GetProp("createdAt")

		updatedFromWallet, err := assets.NewAsset(fromWalletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update source wallet")
		}

		_, err = updatedFromWallet.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving source wallet", err.Status())
		}

		// Save updated destination wallet
		toWalletMap := make(map[string]interface{})
		toWalletMap["@assetType"] = "wallet"
		toWalletMap["@key"] = "wallet:" + toWalletUUID
		toWalletMap["walletId"] = toWalletAsset.GetProp("walletId")
		toWalletMap["ownerPubKey"] = toWalletAsset.GetProp("ownerPubKey")
		toWalletMap["ownerCertHash"] = toWalletAsset.GetProp("ownerCertHash")
		toWalletMap["balances"] = toBalances
		toWalletMap["escrowBalances"] = toEscrowBalances
		toWalletMap["digitalAssetTypes"] = toAssetTypes
		toWalletMap["createdAt"] = toWalletAsset.GetProp("createdAt")

		updatedToWallet, err := assets.NewAsset(toWalletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update destination wallet")
		}

		_, err = updatedToWallet.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving destination wallet", err.Status())
		}

		response := map[string]interface{}{
			"message":      "Transfer completed successfully",
			"fromWalletId": fromWalletUUID,
			"toWalletId":   toWalletUUID,
			"assetId":      assetId,
			"amount":       amount,
		}

		respJSON, jsonErr := json.Marshal(response)
		if jsonErr != nil {
			return nil, errors.WrapError(nil, "failed to encode response to JSON format")
		}

		return respJSON, nil
	},
}

var BurnTokens = transactions.Transaction{
	Tag:         "burnTokens",
	Label:       "Burn Tokens",
	Description: "Burn tokens from a wallet (issuer only)",
	Method:      "POST",
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
			Tag:         "assetId",
			Label:       "Asset ID",
			Description: "ID of the digital asset",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "walletUUID",
			Label:       "Wallet UUID",
			Description: "Wallet to burn tokens from",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "amount",
			Label:       "Amount to Burn",
			Description: "Number of tokens to burn",
			DataType:    "number",
			Required:    true,
		},
		{
			Tag:         "issuerCertHash",
			Label:       "Issuer Certificate Hash",
			Description: "Certificate hash for issuer verification",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		assetId, _ := req["assetId"].(string)
		walletUUID, _ := req["walletUUID"].(string)
		amount, _ := req["amount"].(float64)
		issuerCertHash, _ := req["issuerCertHash"].(string)

		// Verify issuer authorization
		assetKey := assets.Key{"@key": "digitalAsset:" + assetId}
		asset, err := assetKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading digital asset", err.Status())
		}

		if asset.GetProp("issuerHash").(string) != issuerCertHash {
			return nil, errors.NewCCError("Unauthorized: Only asset issuer can burn tokens", 403)
		}

		// Get wallet
		walletKey := assets.Key{"@key": "wallet:" + walletUUID}
		walletAsset, err := walletKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading wallet", err.Status())
		}

		digitalAssetTypes := walletAsset.GetProp("digitalAssetTypes").([]interface{})
		balances := walletAsset.GetProp("balances").([]interface{})

		// Find asset index and update balance
		assetFound := false
		for i, assetRef := range digitalAssetTypes {
			var refAssetId string
			switch ref := assetRef.(type) {
			case map[string]interface{}:
				refAssetId = strings.Split(ref["@key"].(string), ":")[1]
			case string:
				refAssetId = ref
			}

			if refAssetId == assetId {
				currentBalance := balances[i].(float64)
				if currentBalance < amount {
					return nil, errors.NewCCError("Insufficient balance to burn", 400)
				}
				balances[i] = currentBalance - amount
				assetFound = true
				break
			}
		}

		if !assetFound {
			return nil, errors.NewCCError("Asset not found in wallet", 404)
		}

		// Create updated wallet map
		walletMap := make(map[string]interface{})
		walletMap["@assetType"] = "wallet"
		walletMap["@key"] = "wallet:" + walletUUID
		walletMap["walletId"] = walletAsset.GetProp("walletId")
		walletMap["ownerPubKey"] = walletAsset.GetProp("ownerPubKey")
		walletMap["ownerCertHash"] = walletAsset.GetProp("ownerCertHash")
		walletMap["balances"] = balances
		walletMap["escrowBalances"] = walletAsset.GetProp("escrowBalances")
		walletMap["digitalAssetTypes"] = digitalAssetTypes
		walletMap["createdAt"] = walletAsset.GetProp("createdAt")

		updatedWallet, err := assets.NewAsset(walletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update wallet")
		}

		_, err = updatedWallet.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error updating wallet", err.Status())
		}

		// Update total supply
		currentSupply := asset.GetProp("totalSupply").(float64)
		assetMap := make(map[string]interface{})
		assetMap["@assetType"] = "digitalAsset"
		assetMap["@key"] = "digitalAsset:" + assetId
		assetMap["name"] = asset.GetProp("name")
		assetMap["symbol"] = asset.GetProp("symbol")
		assetMap["decimals"] = asset.GetProp("decimals")
		assetMap["totalSupply"] = currentSupply - amount
		assetMap["owner"] = asset.GetProp("owner")
		assetMap["issuedAt"] = asset.GetProp("issuedAt")
		assetMap["issuerHash"] = asset.GetProp("issuerHash")

		updatedAsset, err := assets.NewAsset(assetMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update asset")
		}

		_, err = updatedAsset.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error updating asset", err.Status())
		}

		response := map[string]interface{}{
			"message":     "Tokens burned successfully",
			"assetId":     assetId,
			"walletId":    walletUUID,
			"amount":      amount,
			"totalSupply": currentSupply - amount,
		}

		respJSON, jsonErr := json.Marshal(response)
		if jsonErr != nil {
			return nil, errors.WrapError(nil, "failed to encode response to JSON format")
		}

		return respJSON, nil
	},
}
