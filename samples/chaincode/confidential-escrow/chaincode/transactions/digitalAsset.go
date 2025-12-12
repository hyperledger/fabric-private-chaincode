// This file implements transaction handlers for digital asset token lifecycle management.
// It provides operations for creating, reading, minting, transferring, and burning
// confidential digital tokens with issuer-controlled supply management.
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
	"github.com/hyperledger-labs/cc-tools/events"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	"github.com/hyperledger-labs/cc-tools/transactions"
)

// CreateDigitalAsset initializes a new digital asset token type with fixed parameters.
// Only authorized administrators from Org1MSP or Org2MSP can create new token types.
// The issuer's certificate hash is stored for subsequent authorization of mint/burn operations.
//
// Arguments:
//   - name: Human-readable token name (e.g., "US Dollar Token")
//   - symbol: Unique token identifier (e.g., "USDT")
//   - decimals: Number of decimal places for token precision
//   - totalSupply: Initial total supply of tokens
//   - owner: Identity of the token creator
//   - issuedAt: (Optional) Timestamp of token creation, defaults to current time
//   - issuerHash: Certificate hash of the issuer for access control
//
// Returns:
//   - JSON representation of the created digital asset
//   - Error if asset creation or blockchain persistence fails
//
// Security: Only the entity with matching issuerHash can mint or burn these tokens.
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

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
		name, _ := req["name"].(string)
		symbol, _ := req["symbol"].(string)
		decimals, _ := req["decimals"].(float64)
		totalSupply, _ := req["totalSupply"].(float64)
		owner, _ := req["owner"].(string)
		issuerHash, _ := req["issuerHash"].(string)

		assetMap := make(map[string]any)
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

// ReadDigitalAsset retrieves a digital asset token by its unique identifier.
// This operation is read-only and does not modify ledger state.
//
// Arguments:
//   - uuid: Unique identifier of the digital asset to retrieve
//
// Returns:
//   - JSON representation of the digital asset
//   - Error if asset not found or retrieval fails
//
// Note: Consider implementing symbol-based lookup for improved user experience.
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

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
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

// MintTokens creates new token units and adds them to a specified wallet.
// This operation increases both the wallet balance and the token's total supply.
// Only the original issuer (verified by certificate hash) can mint new tokens.
//
// Arguments:
//   - assetId: UUID of the digital asset token type
//   - pubKey: Public key of the recipient wallet owner
//   - amount: Number of tokens to mint
//   - issuerCertHash: Certificate hash of the issuer for authorization
//
// Process Flow:
//  1. Resolve wallet UUID from public key hash via UserDirectory
//  2. Verify issuer authorization against stored issuerHash
//  3. Update or initialize wallet balance for the asset type
//  4. Increment the token's total supply
//
// Returns:
//   - JSON response with minting details and updated total supply
//   - Error if authorization fails, wallet not found, or update fails
//
// Security: Unauthorized minting attempts are rejected with 403 status.
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
			Tag:      "pubKey",
			Label:    "Public Key",
			DataType: "string",
			Required: true,
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

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
		assetId, _ := req["assetId"].(string)
		pubKey, _ := req["pubKey"].(string)
		amount, _ := req["amount"].(float64)
		issuerCertHash, _ := req["issuerCertHash"].(string)

		// Lookup wallet using publicKeyHash property
		hash := sha256.Sum256([]byte(pubKey))
		pubKeyHash := hex.EncodeToString(hash[:])

		userDirKey, err := assets.NewKey(map[string]any{
			"@assetType":    "userdir",
			"publicKeyHash": pubKeyHash,
		})
		if err != nil {
			return nil, errors.NewCCError(fmt.Sprintf("Seller's Key cannot be found from user dir: %v", err), 404)
		}

		userDir, err := userDirKey.Get(stub)
		if err != nil {
			return nil, errors.NewCCError("Buyer wallet not found. Buyer must create wallet first.", 404)
		}
		walletUUID := userDir.GetProp("walletUUID").(string)

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

		digitalAssetTypes := walletAsset.GetProp("digitalAssetTypes").([]any)
		balances := walletAsset.GetProp("balances").([]any)
		escrowBalances := walletAsset.GetProp("escrowBalances").([]any)

		// Find asset index and update balance
		assetFound := false
		for i, assetRef := range digitalAssetTypes {
			var refAssetId string
			switch ref := assetRef.(type) {
			case map[string]any:
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
			digitalAssetTypes = append(digitalAssetTypes, map[string]any{
				"@key": "digitalAsset:" + assetId,
			})
			balances = append(balances, amount)
			escrowBalances = append(escrowBalances, 0.0)
		}

		// Update wallet
		walletUpdate := map[string]any{
			"balances":          balances,
			"escrowBalances":    escrowBalances,
			"digitalAssetTypes": digitalAssetTypes,
		}

		_, err = walletAsset.Update(stub, walletUpdate)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error updating wallet", err.Status())
		}

		// Update total supply
		currentSupply := asset.GetProp("totalSupply").(float64)
		assetUpdate := map[string]any{
			"totalSupply": currentSupply + amount,
		}

		_, err = asset.Update(stub, assetUpdate)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error updating asset", err.Status())
		}

		response := map[string]any{
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

// TransferTokens moves tokens between two wallets with balance validation.
// This operation atomically decrements the source wallet and increments the destination wallet.
// The sender must provide a valid certificate hash matching the source wallet owner.
//
// Arguments:
//   - fromPubKey: Public key of the sender wallet
//   - toPubKey: Public key of the recipient wallet
//   - assetId: UUID of the digital asset to transfer
//   - amount: Number of tokens to transfer
//   - senderCertHash: Certificate hash of the sender for authorization
//
// Process Flow:
//  1. Resolve both wallet UUIDs from public key hashes
//  2. Verify sender authorization
//  3. Validate sufficient available balance (not escrowed)
//  4. Deduct from source wallet
//  5. Add to destination wallet (initialize asset entry if needed)
//  6. Atomically commit both updates
//
// Returns:
//   - JSON response with transfer confirmation details
//   - Error if insufficient balance, authorization fails, or wallets not found
//
// Security: Only the wallet owner can initiate transfers from their wallet.
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
			Tag:         "fromPubKey",
			Label:       "From Public Key",
			Description: "Source Public Key",
			DataType:    "string",
			Required:    true,
		},
		{
			Tag:         "toPubKey",
			Label:       "To Public Key",
			Description: "Destination Pub Key",
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

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
		fromPubKey, _ := req["fromPubKey"].(string)
		toPubKey, _ := req["toPubKey"].(string)
		assetId, _ := req["assetId"].(string)
		amount, _ := req["amount"].(float64)
		senderCertHash, _ := req["senderCertHash"].(string)

		// Lookup wallet using publicKeyHash property
		hash := sha256.Sum256([]byte(fromPubKey))
		pubKeyHash := hex.EncodeToString(hash[:])

		userDirKey, err := assets.NewKey(map[string]any{
			"@assetType":    "userdir",
			"publicKeyHash": pubKeyHash,
		})
		if err != nil {
			return nil, errors.NewCCError(fmt.Sprintf("Seller's Key cannot be found from user dir: %v", err), 404)
		}

		userDir, err := userDirKey.Get(stub)
		if err != nil {
			return nil, errors.NewCCError("Buyer wallet not found. Buyer must create wallet first.", 404)
		}
		fromWalletUUID := userDir.GetProp("walletUUID").(string)

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

		// Lookup wallet using publicKeyHash property
		hash = sha256.Sum256([]byte(toPubKey))
		pubKeyHash = hex.EncodeToString(hash[:])

		userDirKey, err = assets.NewKey(map[string]any{
			"@assetType":    "userdir",
			"publicKeyHash": pubKeyHash,
		})
		if err != nil {
			return nil, errors.NewCCError(fmt.Sprintf("Seller's Key cannot be found from user dir: %v", err), 404)
		}

		userDir, err = userDirKey.Get(stub)
		if err != nil {
			return nil, errors.NewCCError("Buyer wallet not found. Buyer must create wallet first.", 404)
		}
		toWalletUUID := userDir.GetProp("walletUUID").(string)

		// Get destination wallet
		toKey := assets.Key{"@key": "wallet:" + toWalletUUID}
		toWalletAsset, err := toKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error reading destination wallet", err.Status())
		}

		// Update source wallet balance
		fromAssetTypes := fromWalletAsset.GetProp("digitalAssetTypes").([]any)
		fromBalances := fromWalletAsset.GetProp("balances").([]any)
		fromEscrowBalances := fromWalletAsset.GetProp("escrowBalances").([]any)

		fromAssetFound := false
		for i, assetRef := range fromAssetTypes {
			var refAssetId string
			switch ref := assetRef.(type) {
			case map[string]any:
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
		toAssetTypes := toWalletAsset.GetProp("digitalAssetTypes").([]any)
		toBalances := toWalletAsset.GetProp("balances").([]any)
		toEscrowBalances := toWalletAsset.GetProp("escrowBalances").([]any)

		toAssetFound := false
		for i, assetRef := range toAssetTypes {
			var refAssetId string
			switch ref := assetRef.(type) {
			case map[string]any:
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

		// if asset not found, Add asset
		if !toAssetFound {
			toAssetTypes = append(toAssetTypes, map[string]any{
				"@key": "digitalAsset:" + assetId,
			})
			toBalances = append(toBalances, amount)
			toEscrowBalances = append(toEscrowBalances, 0.0)
		}

		// Save updated source wallet
		fromWalletUpdate := map[string]any{
			"balances":          fromBalances,
			"escrowBalances":    fromEscrowBalances,
			"digitalAssetTypes": fromAssetTypes,
		}
		_, err = fromWalletAsset.Update(stub, fromWalletUpdate)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving source wallet", err.Status())
		}

		// Save updated destination wallet
		toWalletUpdate := map[string]any{
			"balances":          toBalances,
			"escrowBalances":    toEscrowBalances,
			"digitalAssetTypes": toAssetTypes,
		}
		_, err = toWalletAsset.Update(stub, toWalletUpdate)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving destination wallet", err.Status())
		}

		response := map[string]any{
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

// BurnTokens permanently removes tokens from circulation.
// This operation decreases both the wallet balance and the token's total supply.
// Only the original issuer can burn tokens, regardless of which wallet holds them.
//
// Arguments:
//   - assetId: UUID of the digital asset token type
//   - pubKey: Public key of the wallet from which to burn tokens
//   - amount: Number of tokens to burn
//   - issuerCertHash: Certificate hash of the issuer for authorization
//
// Process Flow:
//  1. Resolve wallet UUID from public key hash
//  2. Verify issuer authorization
//  3. Validate sufficient balance in target wallet
//  4. Deduct tokens from wallet balance
//  5. Decrement the token's total supply
//
// Returns:
//   - JSON response with burn details and updated total supply
//   - Error if insufficient balance, authorization fails, or asset not found
//
// Security: Only the token issuer can burn tokens. Wallet owners cannot burn their own tokens.
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
			Tag:         "pubKey",
			Label:       "Public Key",
			Description: "Public Key to burn tokens from",
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

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
		assetId, _ := req["assetId"].(string)
		pubKey, _ := req["pubKey"].(string)
		amount, _ := req["amount"].(float64)
		issuerCertHash, _ := req["issuerCertHash"].(string)

		// Lookup wallet using publicKeyHash property
		hash := sha256.Sum256([]byte(pubKey))
		pubKeyHash := hex.EncodeToString(hash[:])

		userDirKey, err := assets.NewKey(map[string]any{
			"@assetType":    "userdir",
			"publicKeyHash": pubKeyHash,
		})
		if err != nil {
			return nil, errors.NewCCError(fmt.Sprintf("Seller's Key cannot be found from user dir: %v", err), 404)
		}

		userDir, err := userDirKey.Get(stub)
		if err != nil {
			return nil, errors.NewCCError("Buyer wallet not found. Buyer must create wallet first.", 404)
		}
		walletUUID := userDir.GetProp("walletUUID").(string)

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

		digitalAssetTypes := walletAsset.GetProp("digitalAssetTypes").([]any)
		balances := walletAsset.GetProp("balances").([]any)

		// Find asset index and update balance
		assetFound := false
		for i, assetRef := range digitalAssetTypes {
			var refAssetId string
			switch ref := assetRef.(type) {
			case map[string]any:
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
		walletUpdate := map[string]any{
			"balances":          balances,
			"digitalAssetTypes": digitalAssetTypes,
		}
		_, err = walletAsset.Update(stub, walletUpdate)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error updating wallet", err.Status())
		}

		// Update total supply
		currentSupply := asset.GetProp("totalSupply").(float64)
		assetUpdate := map[string]any{
			"totalSupply": currentSupply - amount,
		}
		_, err = asset.Update(stub, assetUpdate)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error updating asset", err.Status())
		}

		response := map[string]any{
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
