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

var CreateAndLockEscrow = transactions.Transaction{
	Tag:         "createAndLockEscrow",
	Label:       "Create and Lock Escrow",
	Description: "Creates a new escrow and immediately locks funds",
	Method:      "POST",
	Callers: []accesscontrol.Caller{
		{MSP: "Org1MSP", OU: "admin"},
		{MSP: "Org2MSP", OU: "admin"},
	},
	Args: []transactions.Argument{
		{Tag: "escrowId", Label: "Escrow ID", DataType: "string", Required: true},
		{Tag: "buyerPubKey", Label: "Buyer Public Key", DataType: "string", Required: true},
		{Tag: "sellerPubKey", Label: "Seller Public Key", DataType: "string", Required: true},
		{Tag: "amount", Label: "Escrowed Amount", DataType: "number", Required: true},
		{Tag: "assetType", Label: "Asset Type Reference", DataType: "->digitalAsset", Required: true},
		// {Tag: "conditionValue", Label: "Condition Value", DataType: "string", Required: true},
		{Tag: "parcelId", Label: "Parcel ID", DataType: "string", Required: true}, // ADD THIS
		{Tag: "secret", Label: "Secret Key", DataType: "string", Required: true},
		{Tag: "buyerWalletUUID", Label: "buyer Wallet UUID", DataType: "string", Required: true},
		{Tag: "sellerWalletUUID", Label: "seller Wallet UUID", DataType: "string", Required: true},
		{Tag: "buyerCertHash", Label: "buyer Certificate Hash", DataType: "string", Required: true},
	},
	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		escrowId, _ := req["escrowId"].(string)
		buyerPubKey, _ := req["buyerPubKey"].(string)
		sellerPubKey, _ := req["sellerPubKey"].(string)
		amount, _ := req["amount"].(float64)
		assetType, _ := req["assetType"].(interface{})
		parcelId, _ := req["parcelId"].(string)
		secret, _ := req["secret"].(string)
		// conditionValue, _ := req["conditionValue"].(string)
		buyerWalletUUID, _ := req["buyerWalletUUID"].(string)
		sellerWalletUUID, _ := req["sellerWalletUUID"].(string)
		buyerCertHash, _ := req["buyerCertHash"].(string)

		// Extract assetId from assetType reference
		var assetId string
		assetKey, ok := assetType.(assets.Key)
		if !ok {
			return nil, errors.NewCCError(fmt.Sprintf("Invalid assetType: expected map, got %T", assetType), 400)
		}

		keyStr, exists := assetKey["@key"]
		if !exists {
			return nil, errors.NewCCError("Invalid assetType: @key field not found", 400)
		}

		keyString, ok := keyStr.(string)
		if !ok {
			return nil, errors.NewCCError(fmt.Sprintf("Invalid assetType: @key is not string, got %T", assetKey), 400)
		}

		parts := strings.Split(keyString, ":")
		if len(parts) != 2 {
			return nil, errors.NewCCError("Invalid assetType: @key format incorrect", 400)
		}
		assetId = parts[1]

		// Check for wallet existence
		// hash := sha256.Sum256([]byte(sellerPubKey))
		// sellerPubKeyHash := hex.EncodeToString(hash[:])
		//
		// fmt.Printf("DEBUG: Seller PubKey: %s\n", sellerPubKey)
		// fmt.Printf("DEBUG: Seller PubKey Hash: %s\n", sellerPubKeyHash)
		//
		// sellerUserDirKey := assets.Key{
		// 	"@assetType":    "userdir",
		// 	"publicKeyHash": sellerPubKeyHash,
		// }
		//
		// fmt.Printf("DEBUG: Attempting to get UserDir with key: %+v\n", sellerUserDirKey)
		//
		// sellerUserDir, err := sellerUserDirKey.Get(stub)
		// if err != nil {
		// 	return nil, errors.NewCCError(fmt.Sprintf("Seller wallet not found. Seller must create wallet first. Details: %v", err), 404)
		// }
		// fmt.Printf("DEBUG: Seller UserDir found: %+v\n", sellerUserDir)
		// sellerWalletUUID := sellerUserDir.GetProp("walletUUID").(string)
		// fmt.Printf("DEBUG: Seller WalletID: %s\n", sellerWalletUUID)
		//
		// // Lookup buyer wallet using publicKeyHash property
		// hash = sha256.Sum256([]byte(buyerPubKey))
		// buyerPubKeyHash := hex.EncodeToString(hash[:])
		//
		// buyerUserDirKey := assets.Key{
		// 	"@assetType":    "userdir",
		// 	"publicKeyHash": buyerPubKeyHash,
		// }
		// buyerUserDir, err := buyerUserDirKey.Get(stub)
		// if err != nil {
		// 	return nil, errors.NewCCError("Buyer wallet not found. Buyer must create wallet first.", 404)
		// }
		// buyerWalletUUID := buyerUserDir.GetProp("walletUUID").(string)
		//
		// fmt.Printf("DEBUG: Seller PubKey received: %s\n", sellerPubKey)
		// hash := sha256.Sum256([]byte(sellerPubKey))
		// sellerPubKeyHash := hex.EncodeToString(hash[:])
		// fmt.Printf("DEBUG: Seller PubKey Hash computed: %s\n", sellerPubKeyHash)
		// fmt.Printf("DEBUG: Looking up key: userdir:%s\n", sellerPubKeyHash)
		//
		// sellerUserDirKey := assets.Key{"@key": "userdir:" + sellerPubKeyHash}
		// sellerUserDir, err := sellerUserDirKey.Get(stub)
		// if err != nil {
		// 	return nil, errors.NewCCError("Seller wallet not found. Seller must create wallet first.", 404)
		// }
		// sellerWalletUUID := sellerUserDir.GetProp("walletUUID").(string)
		//
		// hash = sha256.Sum256([]byte(buyerPubKey))
		// buyerPubKeyHash := hex.EncodeToString(hash[:])
		//
		// buyerUserDirKey := assets.Key{"@key": "userdir:" + buyerPubKeyHash}
		// buyerUserDir, err := buyerUserDirKey.Get(stub)
		// if err != nil {
		// 	return nil, errors.NewCCError("buyer wallet not found. buyer must create wallet first.", 404)
		// }
		// buyerWalletUUID := buyerUserDir.GetProp("walletUUID").(string)

		// 1. Get and verify buyer wallet ownership
		buyerWalletKey := assets.Key{"@key": "wallet:" + buyerWalletUUID}
		buyerWallet, err := buyerWalletKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading buyer wallet", err.Status())
		}

		if buyerWallet.GetProp("ownerCertHash").(string) != buyerCertHash {
			return nil, errors.NewCCError("Unauthorized: Certificate hash mismatch", 403)
		}

		// 2. Get wallet balances
		digitalAssetTypes := buyerWallet.GetProp("digitalAssetTypes").([]interface{})
		balances := buyerWallet.GetProp("balances").([]interface{})

		var escrowBalances []interface{}
		if buyerWallet.GetProp("escrowBalances") != nil {
			escrowBalances = buyerWallet.GetProp("escrowBalances").([]interface{})
		} else {
			escrowBalances = make([]interface{}, len(balances))
			for i := range escrowBalances {
				escrowBalances[i] = 0.0
			}
		}

		// 3. Find asset index and check sufficient balance
		assetFound := false
		assetIndex := -1
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
					return nil, errors.NewCCError("Insufficient balance", 400)
				}
				assetFound = true
				assetIndex = i
				break
			}
		}

		if !assetFound {
			return nil, errors.NewCCError("Asset not found in wallet", 404)
		}

		// 4. Move funds from balances to escrowBalances
		currentBalance := balances[assetIndex].(float64)
		currentEscrowBalance := escrowBalances[assetIndex].(float64)

		balances[assetIndex] = currentBalance - amount
		escrowBalances[assetIndex] = currentEscrowBalance + amount

		// 5. Update wallet
		walletMap := make(map[string]interface{})
		walletMap["@assetType"] = "wallet"
		walletMap["@key"] = "wallet:" + buyerWalletUUID
		walletMap["walletId"] = buyerWallet.GetProp("walletId")
		walletMap["ownerPubKey"] = buyerWallet.GetProp("ownerPubKey")
		walletMap["ownerCertHash"] = buyerWallet.GetProp("ownerCertHash")
		walletMap["balances"] = balances
		walletMap["escrowBalances"] = escrowBalances
		walletMap["digitalAssetTypes"] = digitalAssetTypes
		walletMap["createdAt"] = buyerWallet.GetProp("createdAt")

		updatedWallet, err := assets.NewAsset(walletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update wallet")
		}

		_, err = updatedWallet.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving updated wallet", err.Status())
		}

		// Compute condition hash: SHA256(secret + parcelId)
		conditionData := secret + parcelId
		conditionHash := sha256.Sum256([]byte(conditionData))
		conditionValue := hex.EncodeToString(conditionHash[:])

		// 6. Create escrow with "Active" status
		escrowMap := make(map[string]interface{})
		escrowMap["@assetType"] = "escrow"
		escrowMap["escrowId"] = escrowId
		escrowMap["buyerPubKey"] = buyerPubKey
		escrowMap["sellerPubKey"] = sellerPubKey
		escrowMap["buyerWalletUUID"] = buyerWalletUUID
		escrowMap["sellerWalletUUID"] = sellerWalletUUID
		escrowMap["parcelId"] = parcelId
		escrowMap["amount"] = amount
		escrowMap["assetType"] = assetType
		escrowMap["conditionValue"] = conditionValue
		escrowMap["status"] = "Active"
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

		// // 7. Return response
		// response := map[string]interface{}{
		// 	"message":       "Escrow created and funds locked successfully",
		// 	"escrowId":      escrowId,
		// 	"buyerWalletUUID": buyerWalletUUID,
		// 	"amount":        amount,
		// 	"assetId":       assetId,
		// 	"escrowStatus":  "Active",
		// }
		//
		// responseJSON, jsonErr := json.Marshal(response)
		// if jsonErr != nil {
		// 	return nil, errors.WrapError(nil, "failed to encode response to JSON format")
		// }
		//
		// return responseJSON, nil
	},
}

// var CreateEscrow = transactions.Transaction{
// 	Tag:         "createEscrow",
// 	Label:       "Escrow Creation",
// 	Description: "Creates a new escrow",
// 	Method:      "POST",
// 	Callers: []accesscontrol.Caller{
// 		{
// 			MSP: "Org1MSP",
// 			OU:  "admin",
// 		},
// 		{
// 			MSP: "Org2MSP",
// 			OU:  "admin",
// 		},
// 	},
//
// 	Args: []transactions.Argument{
// 		{
// 			Tag:         "escrowId",
// 			Label:       "Escrow ID",
// 			Description: "ID of Escrow",
// 			DataType:    "string",
// 			Required:    true,
// 		},
// 		{
// 			Tag:      "buyerPubKey",
// 			Label:    "Buyer Public Key",
// 			DataType: "string",
// 			Required: true,
// 		},
// 		{
// 			Tag:      "sellerPubKey",
// 			Label:    "Seller Public Key",
// 			DataType: "string",
// 			Required: true,
// 		},
// 		{
// 			Tag:      "amount",
// 			Label:    "Escrowed Amount",
// 			DataType: "number",
// 			Required: true,
// 		},
// 		{
// 			Tag:      "assetType",
// 			Label:    "Asset Type Reference",
// 			DataType: "->digitalAsset",
// 			Required: true,
// 		},
// 		{
// 			Tag:      "conditionValue",
// 			Label:    "Condition Value",
// 			DataType: "string",
// 			Required: true,
// 		},
// 		{
// 			Tag:      "status",
// 			Label:    "Escrow Status",
// 			DataType: "string",
// 			Required: true,
// 		},
// 		{
// 			Tag:      "createdAt",
// 			Label:    "Creation Timestamp",
// 			DataType: "datetime",
// 			Required: false,
// 		},
// 		{
// 			Tag:      "buyerCertHash",
// 			Label:    "Buyer Certificate Hash",
// 			DataType: "string",
// 			Required: true,
// 		},
// 	},
//
// 	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
// 		escrowId, _ := req["escrowId"].(string)
// 		buyerPubKey, _ := req["buyerPubKey"].(string)
// 		sellerPubKey, _ := req["sellerPubKey"].(string)
// 		amount, _ := req["amount"].(float64)
// 		assetType, _ := req["assetType"].(interface{})
// 		conditionValue, _ := req["conditionValue"].(string)
// 		status, _ := req["status"].(string)
// 		buyerCertHash, _ := req["buyerCertHash"].(string)
//
// 		escrowMap := make(map[string]interface{})
// 		escrowMap["@assetType"] = "escrow"
// 		escrowMap["escrowId"] = escrowId
// 		escrowMap["buyerPubKey"] = buyerPubKey
// 		escrowMap["sellerPubKey"] = sellerPubKey
// 		escrowMap["amount"] = amount
// 		escrowMap["assetType"] = assetType
// 		escrowMap["conditionValue"] = conditionValue
// 		escrowMap["status"] = status
// 		escrowMap["createdAt"] = time.Now()
// 		escrowMap["buyerCertHash"] = buyerCertHash
//
// 		escrowAsset, err := assets.NewAsset(escrowMap)
// 		if err != nil {
// 			return nil, errors.WrapError(err, "Failed to create escrow asset")
// 		}
//
// 		_, err = escrowAsset.PutNew(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error saving escrow on blockchain", err.Status())
// 		}
//
// 		assetJSON, nerr := json.Marshal(escrowAsset)
// 		if nerr != nil {
// 			return nil, errors.WrapError(nil, "failed to encode escrow to JSON format")
// 		}
//
// 		return assetJSON, nil
// 	},
// }
//
// var LockFundsInEscrow = transactions.Transaction{
// 	Tag:         "lockFundsInEscrow",
// 	Label:       "Lock Funds in Escrow",
// 	Description: "Lock funds from buyer wallet into escrow",
// 	Method:      "POST",
// 	Callers: []accesscontrol.Caller{
// 		{MSP: "Org1MSP", OU: "admin"},
// 		{MSP: "Org2MSP", OU: "admin"},
// 	},
// 	Args: []transactions.Argument{
// 		{Tag: "escrowId", Label: "Escrow ID", DataType: "string", Required: true},
// 		{Tag: "buyerWalletUUID", Label: "buyer Wallet ID", DataType: "string", Required: true},
// 		{Tag: "amount", Label: "Amount to Lock", DataType: "number", Required: true},
// 		{Tag: "assetId", Label: "Asset ID", DataType: "string", Required: true},
// 		{Tag: "buyerCertHash", Label: "buyer Certificate Hash", DataType: "string", Required: true},
// 	},
// 	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
// 		escrowId, _ := req["escrowId"].(string)
// 		buyerWalletUUID, _ := req["buyerWalletUUID"].(string)
// 		amount, _ := req["amount"].(float64)
// 		assetId, _ := req["assetId"].(string)
// 		buyerCertHash, _ := req["buyerCertHash"].(string)
//
// 		// 1. Get and verify buyer wallet ownership
// 		buyerWalletKey := assets.Key{"@key": "wallet:" + buyerWalletUUID}
// 		buyerWallet, err := buyerWalletKey.Get(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error reading buyer wallet", err.Status())
// 		}
//
// 		// Verify ownership
// 		if buyerWallet.GetProp("ownerCertHash").(string) != buyerCertHash {
// 			return nil, errors.NewCCError("Unauthorized: Certificate hash mismatch", 403)
// 		}
//
// 		// 2. Get wallet balances
// 		digitalAssetTypes := buyerWallet.GetProp("digitalAssetTypes").([]interface{})
// 		balances := buyerWallet.GetProp("balances").([]interface{})
//
// 		// Get or initialize escrow balances
// 		var escrowBalances []interface{}
// 		if buyerWallet.GetProp("escrowBalances") != nil {
// 			escrowBalances = buyerWallet.GetProp("escrowBalances").([]interface{})
// 		} else {
// 			// Initialize escrow balances if not present
// 			escrowBalances = make([]interface{}, len(balances))
// 			for i := range escrowBalances {
// 				escrowBalances[i] = 0.0
// 			}
// 		}
//
// 		// 3. Find asset index and check sufficient balance
// 		assetFound := false
// 		assetIndex := -1
// 		for i, assetRef := range digitalAssetTypes {
// 			var refAssetId string
// 			switch ref := assetRef.(type) {
// 			case map[string]interface{}:
// 				refAssetId = strings.Split(ref["@key"].(string), ":")[1]
// 			case string:
// 				refAssetId = ref
// 			}
//
// 			if refAssetId == assetId {
// 				currentBalance := balances[i].(float64)
// 				fmt.Println(currentBalance)
// 				fmt.Println(amount)
// 				if currentBalance < amount {
// 					return nil, errors.NewCCError("Insufficient balance", 400)
// 				}
// 				assetFound = true
// 				assetIndex = i
// 				break
// 			}
// 		}
//
// 		if !assetFound {
// 			return nil, errors.NewCCError("Asset not found in wallet", 404)
// 		}
//
// 		// 4. Move funds from balances to escrowBalances
// 		currentBalance := balances[assetIndex].(float64)
// 		currentEscrowBalance := escrowBalances[assetIndex].(float64)
//
// 		balances[assetIndex] = currentBalance - amount
// 		escrowBalances[assetIndex] = currentEscrowBalance + amount
//
// 		// 5. Update wallet
// 		walletMap := make(map[string]interface{})
// 		walletMap["@assetType"] = "wallet"
// 		walletMap["@key"] = "wallet:" + buyerWalletUUID
// 		walletMap["walletId"] = buyerWallet.GetProp("walletId")
// 		walletMap["ownerPubKey"] = buyerWallet.GetProp("ownerPubKey")
// 		walletMap["ownerCertHash"] = buyerWallet.GetProp("ownerCertHash")
// 		walletMap["balances"] = balances
// 		walletMap["escrowBalances"] = escrowBalances
// 		walletMap["digitalAssetTypes"] = digitalAssetTypes
// 		walletMap["createdAt"] = buyerWallet.GetProp("createdAt")
//
// 		updatedWallet, err := assets.NewAsset(walletMap)
// 		if err != nil {
// 			return nil, errors.WrapError(err, "Failed to update wallet")
// 		}
//
// 		_, err = updatedWallet.Put(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error saving updated wallet", err.Status())
// 		}
//
// 		// 6. Update escrow status to "Active"
// 		escrowKey := assets.Key{"@key": "escrow:" + escrowId}
// 		escrowAsset, err := escrowKey.Get(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error reading escrow", err.Status())
// 		}
//
// 		escrowMap := make(map[string]interface{})
// 		escrowMap["@assetType"] = "escrow"
// 		escrowMap["@key"] = "escrow:" + escrowId
// 		escrowMap["escrowId"] = escrowAsset.GetProp("escrowId")
// 		escrowMap["buyerPubKey"] = escrowAsset.GetProp("buyerPubKey")
// 		escrowMap["sellerPubKey"] = escrowAsset.GetProp("sellerPubKey")
// 		escrowMap["amount"] = escrowAsset.GetProp("amount")
// 		escrowMap["assetType"] = escrowAsset.GetProp("assetType")
// 		escrowMap["conditionValue"] = escrowAsset.GetProp("conditionValue")
// 		escrowMap["status"] = "Active" // Update status
// 		escrowMap["createdAt"] = escrowAsset.GetProp("createdAt")
// 		escrowMap["buyerCertHash"] = escrowAsset.GetProp("buyerCertHash")
//
// 		updatedEscrow, err := assets.NewAsset(escrowMap)
// 		if err != nil {
// 			return nil, errors.WrapError(err, "Failed to update escrow")
// 		}
//
// 		_, err = updatedEscrow.Put(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error saving updated escrow", err.Status())
// 		}
//
// 		// 7. Return response
// 		response := map[string]interface{}{
// 			"message":       "Funds locked successfully",
// 			"escrowId":      escrowId,
// 			"buyerWalletUUID": buyerWalletUUID,
// 			"amount":        amount,
// 			"assetId":       assetId,
// 			"escrowStatus":  "Active",
// 		}
//
// 		responseJSON, jsonErr := json.Marshal(response)
// 		if jsonErr != nil {
// 			return nil, errors.WrapError(nil, "failed to encode response to JSON format")
// 		}
//
// 		return responseJSON, nil
// 	},
// }

// Add VerifyEscrowCondition transaction
var VerifyEscrowCondition = transactions.Transaction{
	Tag: "verifyEscrowCondition",
	Args: []transactions.Argument{
		{Tag: "escrowId", DataType: "string", Required: true},
		{Tag: "secret", DataType: "string", Required: true},
		{Tag: "parcelId", DataType: "string", Required: true},
	},
	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		escrowId, _ := req["escrowId"].(string)
		secret, _ := req["secret"].(string)
		parcelId, _ := req["parcelId"].(string)

		// 1. Get escrow by ID
		escrowKey := assets.Key{"@key": "escrow:" + escrowId}
		escrowAsset, err := escrowKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading escrow", err.Status())
		}

		// Check escrow status
		currentStatus := escrowAsset.GetProp("status").(string)
		if currentStatus != "Active" {
			return nil, errors.NewCCError("Escrow is not active", 400)
		}

		// 2. Get stored condition value from escrow
		storedCondition := escrowAsset.GetProp("conditionValue").(string)

		// 3. Compute SHA256(secret + parcelId)
		hasher := sha256.New()
		hasher.Write([]byte(secret + parcelId))
		computedHash := hex.EncodeToString(hasher.Sum(nil))

		// 4. Verify condition: sha256(secret + parcelID) == stored condition
		if computedHash != storedCondition {
			return nil, errors.NewCCError("Condition verification failed: hash mismatch", 403)
		}

		// 5. Update escrow status to "ReadyForRelease"
		escrowMap := make(map[string]interface{})
		escrowMap["@assetType"] = "escrow"
		escrowMap["@key"] = "escrow:" + escrowId
		escrowMap["escrowId"] = escrowAsset.GetProp("escrowId")
		escrowMap["buyerPubKey"] = escrowAsset.GetProp("buyerPubKey")
		escrowMap["sellerPubKey"] = escrowAsset.GetProp("sellerPubKey")
		escrowMap["parcelId"] = escrowAsset.GetProp("parcelId")
		escrowMap["amount"] = escrowAsset.GetProp("amount")
		escrowMap["assetType"] = escrowAsset.GetProp("assetType")
		escrowMap["conditionValue"] = escrowAsset.GetProp("conditionValue")
		escrowMap["status"] = "ReadyForRelease" // Update status
		escrowMap["createdAt"] = escrowAsset.GetProp("createdAt")
		escrowMap["buyerCertHash"] = escrowAsset.GetProp("buyerCertHash")

		updatedEscrow, err := assets.NewAsset(escrowMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update escrow")
		}

		_, err = updatedEscrow.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving updated escrow", err.Status())
		}

		// 6. Return success response
		response := map[string]interface{}{
			"message":      "Condition verified successfully",
			"escrowId":     escrowId,
			"status":       "ReadyForRelease",
			"parcelId":     parcelId,
			"computedHash": computedHash,
		}

		responseJSON, jsonErr := json.Marshal(response)
		if jsonErr != nil {
			return nil, errors.WrapError(nil, "failed to encode response to JSON format")
		}

		return responseJSON, nil
	},
}

var ReleaseEscrow = transactions.Transaction{
	Tag:         "releaseEscrow",
	Label:       "Release Escrow",
	Description: "Seller releases escrow with secret and parcelId",
	Method:      "POST",
	Callers: []accesscontrol.Caller{
		{MSP: "Org1MSP", OU: "admin"},
		{MSP: "Org2MSP", OU: "admin"},
	},
	Args: []transactions.Argument{
		{Tag: "escrowUUID", DataType: "string", Required: true},
		{Tag: "secret", DataType: "string", Required: true},
		{Tag: "parcelId", DataType: "string", Required: true},
		{Tag: "sellerCertHash", DataType: "string", Required: true},
	},
	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		escrowUUID, _ := req["escrowUUID"].(string)
		secret, _ := req["secret"].(string)
		parcelId, _ := req["parcelId"].(string)
		sellerCertHash, _ := req["sellerCertHash"].(string)

		// Get escrow
		escrowKey := assets.Key{"@key": "escrow:" + escrowUUID}
		escrowAsset, err := escrowKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Escrow not found", 404)
		}

		// Verify status
		if escrowAsset.GetProp("status").(string) != "Active" {
			return nil, errors.NewCCError("Escrow is not active", 400)
		}

		// Verify parcelId matches
		if escrowAsset.GetProp("parcelId").(string) != parcelId {
			return nil, errors.NewCCError("Invalid parcel ID", 403)
		}

		// Verify condition: SHA256(secret + parcelId)
		conditionData := secret + parcelId
		computedHash := sha256.Sum256([]byte(conditionData))
		computedCondition := hex.EncodeToString(computedHash[:])

		storedCondition := escrowAsset.GetProp("conditionValue").(string)
		if computedCondition != storedCondition {
			return nil, errors.NewCCError("Invalid secret", 403)
		}

		// Get seller wallet
		sellerWalletId := escrowAsset.GetProp("sellerWalletUUID").(string)
		sellerWalletKey := assets.Key{"@key": "wallet:" + sellerWalletId}
		sellerWallet, err := sellerWalletKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Seller wallet not found", 404)
		}

		// Verify seller authorization
		if sellerWallet.GetProp("ownerCertHash").(string) != sellerCertHash {
			return nil, errors.NewCCError("Unauthorized: Not the seller", 403)
		}

		// Get buyer wallet
		buyerWalletId := escrowAsset.GetProp("buyerWalletUUID").(string)
		buyerWalletKey := assets.Key{"@key": "wallet:" + buyerWalletId}
		buyerWallet, err := buyerWalletKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Buyer wallet not found", 404)
		}

		// Get asset info
		assetType := escrowAsset.GetProp("assetType").(map[string]interface{})
		assetId := strings.Split(assetType["@key"].(string), ":")[1]
		amount := escrowAsset.GetProp("amount").(float64)

		// Find asset index in both wallets
		buyerAssets := buyerWallet.GetProp("digitalAssetTypes").([]interface{})
		buyerBalances := buyerWallet.GetProp("balances").([]interface{})
		buyerEscrowBalances := buyerWallet.GetProp("escrowBalances").([]interface{})

		sellerAssets := sellerWallet.GetProp("digitalAssetTypes").([]interface{})
		sellerBalances := sellerWallet.GetProp("balances").([]interface{})

		var sellerEscrowBalances []interface{}
		if sellerWallet.GetProp("escrowBalances") != nil {
			sellerEscrowBalances = sellerWallet.GetProp("escrowBalances").([]interface{})
		} else {
			sellerEscrowBalances = make([]interface{}, len(sellerBalances))
			for i := range sellerEscrowBalances {
				sellerEscrowBalances[i] = 0.0
			}
		}

		var buyerAssetIndex, sellerAssetIndex int = -1, -1

		// Find buyer asset index
		for i, assetRef := range buyerAssets {
			refAssetId := strings.Split(assetRef.(map[string]interface{})["@key"].(string), ":")[1]
			if refAssetId == assetId {
				buyerAssetIndex = i
				break
			}
		}

		// Find seller asset index
		for i, assetRef := range sellerAssets {
			refAssetId := strings.Split(assetRef.(map[string]interface{})["@key"].(string), ":")[1]
			if refAssetId == assetId {
				sellerAssetIndex = i
				break
			}
		}

		// if buyerAssetIndex == -1 || sellerAssetIndex == -1 {
		// 	return nil, errors.NewCCError("Asset not found in wallets", 404)
		// }

		if sellerAssetIndex == -1 {
			sellerAssets = append(sellerAssets, assetType)
			sellerBalances = append(sellerBalances, 0.0)
			sellerEscrowBalances = append(sellerEscrowBalances, 0.0)
			sellerAssetIndex = len(sellerAssets) - 1
		}

		// Transfer: Reduce buyer escrow balance, increase seller balance
		buyerEscrowBalances[buyerAssetIndex] = buyerEscrowBalances[buyerAssetIndex].(float64) - amount
		sellerBalances[sellerAssetIndex] = sellerBalances[sellerAssetIndex].(float64) + amount

		// Update buyer wallet
		buyerWalletMap := make(map[string]interface{})
		buyerWalletMap["@assetType"] = "wallet"
		buyerWalletMap["@key"] = "wallet:" + buyerWalletId
		buyerWalletMap["walletId"] = buyerWallet.GetProp("walletId")
		buyerWalletMap["ownerPubKey"] = buyerWallet.GetProp("ownerPubKey")
		buyerWalletMap["ownerCertHash"] = buyerWallet.GetProp("ownerCertHash")
		buyerWalletMap["balances"] = buyerBalances
		buyerWalletMap["escrowBalances"] = buyerEscrowBalances
		buyerWalletMap["digitalAssetTypes"] = buyerAssets
		buyerWalletMap["createdAt"] = buyerWallet.GetProp("createdAt")

		updatedBuyerWallet, err := assets.NewAsset(buyerWalletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update buyer wallet")
		}
		_, err = updatedBuyerWallet.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Failed to save buyer wallet", err.Status())
		}

		// Update seller wallet
		sellerWalletMap := make(map[string]interface{})
		sellerWalletMap["@assetType"] = "wallet"
		sellerWalletMap["@key"] = "wallet:" + sellerWalletId
		sellerWalletMap["walletId"] = sellerWallet.GetProp("walletId")
		sellerWalletMap["ownerPubKey"] = sellerWallet.GetProp("ownerPubKey")
		sellerWalletMap["ownerCertHash"] = sellerWallet.GetProp("ownerCertHash")
		sellerWalletMap["balances"] = sellerBalances
		sellerWalletMap["escrowBalances"] = sellerEscrowBalances
		// sellerWalletMap["escrowBalances"] = sellerWallet.GetProp("escrowBalances")
		sellerWalletMap["digitalAssetTypes"] = sellerAssets
		sellerWalletMap["createdAt"] = sellerWallet.GetProp("createdAt")

		updatedSellerWallet, err := assets.NewAsset(sellerWalletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update seller wallet")
		}
		_, err = updatedSellerWallet.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Failed to save seller wallet", err.Status())
		}

		// Update escrow status to Released
		escrowMap := make(map[string]interface{})
		escrowMap["@assetType"] = "escrow"
		escrowMap["@key"] = "escrow:" + escrowUUID
		escrowMap["escrowId"] = escrowAsset.GetProp("escrowId")
		escrowMap["buyerPubKey"] = escrowAsset.GetProp("buyerPubKey")
		escrowMap["sellerPubKey"] = escrowAsset.GetProp("sellerPubKey")
		escrowMap["buyerWalletUUID"] = buyerWalletId
		escrowMap["sellerWalletUUID"] = sellerWalletId
		escrowMap["amount"] = amount
		escrowMap["assetType"] = assetType
		escrowMap["parcelId"] = parcelId
		escrowMap["conditionValue"] = storedCondition
		escrowMap["status"] = "Released"
		escrowMap["createdAt"] = escrowAsset.GetProp("createdAt")
		escrowMap["buyerCertHash"] = escrowAsset.GetProp("buyerCertHash")

		updatedEscrow, err := assets.NewAsset(escrowMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update escrow")
		}
		_, err = updatedEscrow.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Failed to save escrow", err.Status())
		}

		response := map[string]interface{}{
			"message":        "Escrow released successfully",
			"escrowId":       escrowUUID,
			"amount":         amount,
			"sellerWalletId": sellerWalletId,
		}

		responseJSON, _ := json.Marshal(response)
		return responseJSON, nil
	},
}

// var ReleaseEscrow = transactions.Transaction{
// 	Tag: "releaseEscrow",
// 	Args: []transactions.Argument{
// 		{Tag: "escrowUUID", DataType: "string", Required: true},
// 		{Tag: "secret", DataType: "string", Required: true},
// 		{Tag: "parcelId", DataType: "string", Required: true},
// 		{Tag: "buyerWalletUUID", Label: "buyer Wallet UUID", DataType: "string", Required: true},
// 	},
// 	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
// 		escrowId, _ := req["escrowId"].(string)
// 		buyerWalletUUID, _ := req["buyerWalletUUID"].(string)
//
// 		// 1. Verify escrow status is "ReadyForRelease"
// 		escrowKey := assets.Key{"@key": "escrow:" + escrowId}
// 		escrowAsset, err := escrowKey.Get(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error reading escrow", err.Status())
// 		}
//
// 		currentStatus := escrowAsset.GetProp("status").(string)
// 		if currentStatus != "ReadyForRelease" {
// 			return nil, errors.NewCCError("Escrow is not ready for release", 400)
// 		}
//
// 		escrowAmount := escrowAsset.GetProp("amount").(float64)
//
// 		// Get seller wallet ID from escrow (stored during creation)
// 		sellerWalletUUID := escrowAsset.GetProp("sellerWalletUUID").(string)
//
// 		// Get asset reference from escrow
// 		assetTypeRef := escrowAsset.GetProp("assetType")
// 		var assetId string
// 		switch ref := assetTypeRef.(type) {
// 		case map[string]interface{}:
// 			assetId = strings.Split(ref["@key"].(string), ":")[1]
// 		case string:
// 			assetId = ref
// 		}
//
// 		// 2. Get buyer wallet
// 		buyerWalletKey := assets.Key{"@key": "wallet:" + buyerWalletUUID}
// 		buyerWallet, err := buyerWalletKey.Get(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error reading buyer wallet", err.Status())
// 		}
//
// 		// 3. Get seller wallet
// 		sellerWalletKey := assets.Key{"@key": "wallet:" + sellerWalletUUID}
// 		sellerWallet, err := sellerWalletKey.Get(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error reading seller wallet", err.Status())
// 		}
//
// 		// 4. Move funds from buyer's escrowBalances to seller's balances
// 		// Update buyer wallet
// 		buyerAssetTypes := buyerWallet.GetProp("digitalAssetTypes").([]interface{})
// 		buyerBalances := buyerWallet.GetProp("balances").([]interface{})
// 		buyerEscrowBalances := buyerWallet.GetProp("escrowBalances").([]interface{})
//
// 		// Find asset in buyer wallet and reduce escrow balance
// 		buyerAssetFound := false
// 		for i, assetRef := range buyerAssetTypes {
// 			var refAssetId string
// 			switch ref := assetRef.(type) {
// 			case map[string]interface{}:
// 				refAssetId = strings.Split(ref["@key"].(string), ":")[1]
// 			case string:
// 				refAssetId = ref
// 			}
//
// 			if refAssetId == assetId {
// 				currentEscrowBalance := buyerEscrowBalances[i].(float64)
// 				if currentEscrowBalance < escrowAmount {
// 					return nil, errors.NewCCError("Insufficient escrow balance", 400)
// 				}
// 				buyerEscrowBalances[i] = currentEscrowBalance - escrowAmount
// 				buyerAssetFound = true
// 				break
// 			}
// 		}
//
// 		if !buyerAssetFound {
// 			return nil, errors.NewCCError("Asset not found in buyer wallet", 404)
// 		}
//
// 		// Update seller wallet
// 		sellerAssetTypes := sellerWallet.GetProp("digitalAssetTypes").([]interface{})
// 		sellerBalances := sellerWallet.GetProp("balances").([]interface{})
//
// 		// Initialize seller escrow balances if not present
// 		var sellerEscrowBalances []interface{}
// 		if sellerWallet.GetProp("escrowBalances") != nil {
// 			sellerEscrowBalances = sellerWallet.GetProp("escrowBalances").([]interface{})
// 		} else {
// 			sellerEscrowBalances = make([]interface{}, len(sellerBalances))
// 			for i := range sellerEscrowBalances {
// 				sellerEscrowBalances[i] = 0.0
// 			}
// 		}
//
// 		// Find asset in seller wallet and increase balance
// 		sellerAssetFound := false
// 		for i, assetRef := range sellerAssetTypes {
// 			var refAssetId string
// 			switch ref := assetRef.(type) {
// 			case map[string]interface{}:
// 				refAssetId = strings.Split(ref["@key"].(string), ":")[1]
// 			case string:
// 				refAssetId = ref
// 			}
//
// 			if refAssetId == assetId {
// 				currentBalance := sellerBalances[i].(float64)
// 				sellerBalances[i] = currentBalance + escrowAmount
// 				sellerAssetFound = true
// 				break
// 			}
// 		}
//
// 		// If asset not found in seller wallet, add it
// 		if !sellerAssetFound {
// 			sellerAssetTypes = append(sellerAssetTypes, map[string]interface{}{
// 				"@key": "digitalAsset:" + assetId,
// 			})
// 			sellerBalances = append(sellerBalances, escrowAmount)
// 			sellerEscrowBalances = append(sellerEscrowBalances, 0.0)
// 		}
//
// 		// 5. Save updated buyer wallet
// 		buyerWalletMap := make(map[string]interface{})
// 		buyerWalletMap["@assetType"] = "wallet"
// 		buyerWalletMap["@key"] = "wallet:" + buyerWalletUUID
// 		buyerWalletMap["walletId"] = buyerWallet.GetProp("walletId")
// 		buyerWalletMap["ownerPubKey"] = buyerWallet.GetProp("ownerPubKey")
// 		buyerWalletMap["ownerCertHash"] = buyerWallet.GetProp("ownerCertHash")
// 		buyerWalletMap["balances"] = buyerBalances
// 		buyerWalletMap["escrowBalances"] = buyerEscrowBalances
// 		buyerWalletMap["digitalAssetTypes"] = buyerAssetTypes
// 		buyerWalletMap["createdAt"] = buyerWallet.GetProp("createdAt")
//
// 		updatedbuyerWallet, err := assets.NewAsset(buyerWalletMap)
// 		if err != nil {
// 			return nil, errors.WrapError(err, "Failed to update buyer wallet")
// 		}
//
// 		_, err = updatedbuyerWallet.Put(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error saving buyer wallet", err.Status())
// 		}
//
// 		// Save updated seller wallet
// 		sellerWalletMap := make(map[string]interface{})
// 		sellerWalletMap["@assetType"] = "wallet"
// 		sellerWalletMap["@key"] = "wallet:" + sellerWalletUUID
// 		sellerWalletMap["walletId"] = sellerWallet.GetProp("walletId")
// 		sellerWalletMap["ownerPubKey"] = sellerWallet.GetProp("ownerPubKey")
// 		sellerWalletMap["ownerCertHash"] = sellerWallet.GetProp("ownerCertHash")
// 		sellerWalletMap["balances"] = sellerBalances
// 		sellerWalletMap["escrowBalances"] = sellerEscrowBalances
// 		sellerWalletMap["digitalAssetTypes"] = sellerAssetTypes
// 		sellerWalletMap["createdAt"] = sellerWallet.GetProp("createdAt")
//
// 		updatedsellerWallet, err := assets.NewAsset(sellerWalletMap)
// 		if err != nil {
// 			return nil, errors.WrapError(err, "Failed to update seller wallet")
// 		}
//
// 		_, err = updatedsellerWallet.Put(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error saving seller wallet", err.Status())
// 		}
//
// 		// Update escrow status to "Released"
// 		escrowMap := make(map[string]interface{})
// 		escrowMap["@assetType"] = "escrow"
// 		escrowMap["@key"] = "escrow:" + escrowId
// 		escrowMap["escrowId"] = escrowAsset.GetProp("escrowId")
// 		escrowMap["buyerPubKey"] = escrowAsset.GetProp("buyerPubKey")
// 		escrowMap["sellerPubKey"] = escrowAsset.GetProp("sellerPubKey")
// 		escrowMap["amount"] = escrowAsset.GetProp("amount")
// 		escrowMap["assetType"] = escrowAsset.GetProp("assetType")
// 		escrowMap["conditionValue"] = escrowAsset.GetProp("conditionValue")
// 		escrowMap["status"] = "Released"
// 		escrowMap["createdAt"] = escrowAsset.GetProp("createdAt")
// 		escrowMap["buyerCertHash"] = escrowAsset.GetProp("buyerCertHash")
//
// 		updatedEscrow, err := assets.NewAsset(escrowMap)
// 		if err != nil {
// 			return nil, errors.WrapError(err, "Failed to update escrow")
// 		}
//
// 		_, err = updatedEscrow.Put(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error saving updated escrow", err.Status())
// 		}
//
// 		// Return response
// 		response := map[string]interface{}{
// 			"message":          "Escrow funds released successfully",
// 			"escrowId":         escrowId,
// 			"buyerWalletUUID":  buyerWalletUUID,
// 			"sellerWalletUUID": sellerWalletUUID,
// 			"amount":           escrowAmount,
// 			"assetId":          assetId,
// 			"status":           "Released",
// 		}
//
// 		responseJSON, jsonErr := json.Marshal(response)
// 		if jsonErr != nil {
// 			return nil, errors.WrapError(nil, "failed to encode response to JSON format")
// 		}
//
// 		return responseJSON, nil
// 	},
// }

var RefundEscrow = transactions.Transaction{
	Tag:         "refundEscrow",
	Label:       "Refund Escrow",
	Description: "Buyer refunds escrow if condition not met",
	Method:      "POST",
	Callers: []accesscontrol.Caller{
		{MSP: "Org1MSP", OU: "admin"},
		{MSP: "Org2MSP", OU: "admin"},
	},
	Args: []transactions.Argument{
		{Tag: "escrowUUID", DataType: "string", Required: true},
		{Tag: "buyerWalletUUID", DataType: "string", Required: true},
		{Tag: "buyerCertHash", DataType: "string", Required: true},
	},
	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		escrowUUID, _ := req["escrowUUID"].(string)
		buyerWalletUUID, _ := req["buyerWalletUUID"].(string)
		buyerCertHash, _ := req["buyerCertHash"].(string)

		// Get escrow
		escrowKey := assets.Key{"@key": "escrow:" + escrowUUID}
		escrowAsset, err := escrowKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Escrow not found", 404)
		}

		// Verify status
		if escrowAsset.GetProp("status").(string) != "Active" {
			return nil, errors.NewCCError("Escrow is not active", 400)
		}

		// Get buyer wallet
		// buyerWalletId := escrowAsset.GetProp("buyerWalletId").(string)
		// buyerWalletKey := assets.Key{"@key": "wallet:" + buyerWalletId}
		// buyerWallet, err := buyerWalletKey.Get(stub)
		// if err != nil {
		// 	return nil, errors.WrapErrorWithStatus(err, "Buyer wallet not found", 404)
		// }
		//
		// // Verify buyer authorization
		// if buyerWallet.GetProp("ownerCertHash").(string) != buyerCertHash {
		// 	return nil, errors.NewCCError("Unauthorized: Not the buyer", 403)
		// }
		buyerWalletKey := assets.Key{"@key": "wallet:" + buyerWalletUUID} // CHANGED
		buyerWallet, err := buyerWalletKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Buyer wallet not found", 404)
		}
		if buyerWallet.GetProp("ownerCertHash").(string) != buyerCertHash {
			return nil, errors.NewCCError("Unauthorized: Not the buyer", 403)
		}

		// Get asset info
		assetType := escrowAsset.GetProp("assetType").(map[string]interface{})
		assetId := strings.Split(assetType["@key"].(string), ":")[1]
		amount := escrowAsset.GetProp("amount").(float64)

		// Find asset index
		buyerAssets := buyerWallet.GetProp("digitalAssetTypes").([]interface{})
		buyerBalances := buyerWallet.GetProp("balances").([]interface{})
		buyerEscrowBalances := buyerWallet.GetProp("escrowBalances").([]interface{})

		var buyerAssetIndex int = -1
		for i, assetRef := range buyerAssets {
			refAssetId := strings.Split(assetRef.(map[string]interface{})["@key"].(string), ":")[1]
			if refAssetId == assetId {
				buyerAssetIndex = i
				break
			}
		}

		if buyerAssetIndex == -1 {
			return nil, errors.NewCCError("Asset not found in wallet", 404)
		}

		// Refund: Move from escrow back to available balance
		buyerEscrowBalances[buyerAssetIndex] = buyerEscrowBalances[buyerAssetIndex].(float64) - amount
		buyerBalances[buyerAssetIndex] = buyerBalances[buyerAssetIndex].(float64) + amount

		// Update buyer wallet
		buyerWalletMap := make(map[string]interface{})
		buyerWalletMap["@assetType"] = "wallet"
		buyerWalletMap["@key"] = "wallet:" + buyerWalletUUID
		buyerWalletMap["walletId"] = buyerWallet.GetProp("walletId")
		buyerWalletMap["ownerPubKey"] = buyerWallet.GetProp("ownerPubKey")
		buyerWalletMap["ownerCertHash"] = buyerWallet.GetProp("ownerCertHash")
		buyerWalletMap["balances"] = buyerBalances
		buyerWalletMap["escrowBalances"] = buyerEscrowBalances
		buyerWalletMap["digitalAssetTypes"] = buyerAssets
		buyerWalletMap["createdAt"] = buyerWallet.GetProp("createdAt")

		updatedBuyerWallet, err := assets.NewAsset(buyerWalletMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update buyer wallet")
		}
		_, err = updatedBuyerWallet.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Failed to save buyer wallet", err.Status())
		}

		// Update escrow status to Refunded
		escrowMap := make(map[string]interface{})
		escrowMap["@assetType"] = "escrow"
		escrowMap["@key"] = "escrow:" + escrowUUID
		escrowMap["escrowId"] = escrowAsset.GetProp("escrowId")
		escrowMap["buyerPubKey"] = escrowAsset.GetProp("buyerPubKey")
		escrowMap["sellerPubKey"] = escrowAsset.GetProp("sellerPubKey")
		escrowMap["buyerWalletUUID"] = buyerWalletUUID
		escrowMap["sellerWalletUUID"] = escrowAsset.GetProp("sellerWalletUUID")
		escrowMap["amount"] = amount
		escrowMap["assetType"] = assetType
		escrowMap["parcelId"] = escrowAsset.GetProp("parcelId")
		escrowMap["conditionValue"] = escrowAsset.GetProp("conditionValue")
		escrowMap["status"] = "Refunded"
		escrowMap["createdAt"] = escrowAsset.GetProp("createdAt")
		escrowMap["buyerCertHash"] = buyerCertHash

		updatedEscrow, err := assets.NewAsset(escrowMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to update escrow")
		}
		_, err = updatedEscrow.Put(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Failed to save escrow", err.Status())
		}

		response := map[string]interface{}{
			"message":         "Escrow refunded successfully",
			"escrowUUID":      escrowUUID,
			"amount":          amount,
			"buyerWalletUUID": buyerWalletUUID,
		}

		responseJSON, _ := json.Marshal(response)
		return responseJSON, nil
	},
}

// var RefundEscrow = transactions.Transaction{
// 	Tag: "refundEscrow",
// 	Args: []transactions.Argument{
// 		{Tag: "escrowId", DataType: "string", Required: true},
// 		{Tag: "buyerWalletUUID", Label: "buyer Wallet UUID", DataType: "string", Required: true},
// 	},
// 	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
// 		escrowId, _ := req["escrowId"].(string)
// 		buyerWalletUUID, _ := req["buyerWalletUUID"].(string)
//
// 		// 1. Get escrow
// 		escrowKey := assets.Key{"@key": "escrow:" + escrowId}
// 		escrowAsset, err := escrowKey.Get(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error reading escrow", err.Status())
// 		}
//
// 		// Check escrow is not already released or refunded
// 		currentStatus := escrowAsset.GetProp("status").(string)
// 		if currentStatus == "Released" || currentStatus == "Refunded" {
// 			return nil, errors.NewCCError("Escrow already "+currentStatus, 400)
// 		}
//
// 		escrowAmount := escrowAsset.GetProp("amount").(float64)
//
// 		// Get asset reference from escrow
// 		assetTypeRef := escrowAsset.GetProp("assetType")
// 		var assetId string
// 		switch ref := assetTypeRef.(type) {
// 		case map[string]interface{}:
// 			assetId = strings.Split(ref["@key"].(string), ":")[1]
// 		case string:
// 			assetId = ref
// 		}
//
// 		// 2. Get buyer wallet
// 		buyerWalletKey := assets.Key{"@key": "wallet:" + buyerWalletUUID}
// 		buyerWallet, err := buyerWalletKey.Get(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error reading buyer wallet", err.Status())
// 		}
//
// 		// 3. Move funds from buyer's escrowBalances back to balances
// 		buyerAssetTypes := buyerWallet.GetProp("digitalAssetTypes").([]interface{})
// 		buyerBalances := buyerWallet.GetProp("balances").([]interface{})
// 		buyerEscrowBalances := buyerWallet.GetProp("escrowBalances").([]interface{})
//
// 		// Find asset in buyer wallet
// 		buyerAssetFound := false
// 		for i, assetRef := range buyerAssetTypes {
// 			var refAssetId string
// 			switch ref := assetRef.(type) {
// 			case map[string]interface{}:
// 				refAssetId = strings.Split(ref["@key"].(string), ":")[1]
// 			case string:
// 				refAssetId = ref
// 			}
//
// 			if refAssetId == assetId {
// 				currentBalance := buyerBalances[i].(float64)
// 				currentEscrowBalance := buyerEscrowBalances[i].(float64)
//
// 				if currentEscrowBalance < escrowAmount {
// 					return nil, errors.NewCCError("Insufficient escrow balance", 400)
// 				}
//
// 				// Move funds from escrow back to available balance
// 				buyerBalances[i] = currentBalance + escrowAmount
// 				buyerEscrowBalances[i] = currentEscrowBalance - escrowAmount
// 				buyerAssetFound = true
// 				break
// 			}
// 		}
//
// 		if !buyerAssetFound {
// 			return nil, errors.NewCCError("Asset not found in buyer wallet", 404)
// 		}
//
// 		// 4. Save updated buyer wallet
// 		buyerWalletMap := make(map[string]interface{})
// 		buyerWalletMap["@assetType"] = "wallet"
// 		buyerWalletMap["@key"] = "wallet:" + buyerWalletUUID
// 		buyerWalletMap["walletId"] = buyerWallet.GetProp("walletId")
// 		buyerWalletMap["ownerPubKey"] = buyerWallet.GetProp("ownerPubKey")
// 		buyerWalletMap["ownerCertHash"] = buyerWallet.GetProp("ownerCertHash")
// 		buyerWalletMap["balances"] = buyerBalances
// 		buyerWalletMap["escrowBalances"] = buyerEscrowBalances
// 		buyerWalletMap["digitalAssetTypes"] = buyerAssetTypes
// 		buyerWalletMap["createdAt"] = buyerWallet.GetProp("createdAt")
//
// 		updatedbuyerWallet, err := assets.NewAsset(buyerWalletMap)
// 		if err != nil {
// 			return nil, errors.WrapError(err, "Failed to update buyer wallet")
// 		}
//
// 		_, err = updatedbuyerWallet.Put(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error saving buyer wallet", err.Status())
// 		}
//
// 		// 5. Update escrow status to "Refunded"
// 		escrowMap := make(map[string]interface{})
// 		escrowMap["@assetType"] = "escrow"
// 		escrowMap["@key"] = "escrow:" + escrowId
// 		escrowMap["escrowId"] = escrowAsset.GetProp("escrowId")
// 		escrowMap["buyerPubKey"] = escrowAsset.GetProp("buyerPubKey")
// 		escrowMap["sellerPubKey"] = escrowAsset.GetProp("sellerPubKey")
// 		escrowMap["amount"] = escrowAsset.GetProp("amount")
// 		escrowMap["assetType"] = escrowAsset.GetProp("assetType")
// 		escrowMap["conditionValue"] = escrowAsset.GetProp("conditionValue")
// 		escrowMap["status"] = "Refunded"
// 		escrowMap["createdAt"] = escrowAsset.GetProp("createdAt")
// 		escrowMap["buyerCertHash"] = escrowAsset.GetProp("buyerCertHash")
//
// 		updatedEscrow, err := assets.NewAsset(escrowMap)
// 		if err != nil {
// 			return nil, errors.WrapError(err, "Failed to update escrow")
// 		}
//
// 		_, err = updatedEscrow.Put(stub)
// 		if err != nil {
// 			return nil, errors.WrapErrorWithStatus(err, "Error saving updated escrow", err.Status())
// 		}
//
// 		// Return response
// 		response := map[string]interface{}{
// 			"message":         "Escrow funds refunded successfully",
// 			"escrowId":        escrowId,
// 			"buyerWalletUUID": buyerWalletUUID,
// 			"amount":          escrowAmount,
// 			"assetId":         assetId,
// 			"status":          "Refunded",
// 		}
//
// 		responseJSON, jsonErr := json.Marshal(response)
// 		if jsonErr != nil {
// 			return nil, errors.WrapError(nil, "failed to encode response to JSON format")
// 		}
//
// 		return responseJSON, nil
// 	},
// }

var ReadEscrow = transactions.Transaction{
	Tag:         "readEscrow",
	Label:       "Read Escrow",
	Description: "Read an Escrow by its escrowId",
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
		// PPS: Cross-check method is needed...
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
			"@key": "escrow:" + uuid,
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
