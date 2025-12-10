// This file implements wallet management operations for confidential digital asset accounts.
// It provides secure wallet creation, balance queries, and ownership verification using
// certificate-based authentication and public key hash lookups.
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

// CreateWallet initializes a new wallet for a user and registers it in the UserDirectory.
// This atomic operation creates both the wallet asset and its corresponding directory entry,
// enabling future wallet lookups by public key hash.
//
// Arguments:
//   - walletId: User-defined identifier for the wallet
//   - ownerPubKey: Public key of the wallet owner
//   - ownerCertHash: Certificate hash for ownership verification
//
// Process Flow:
//  1. Create wallet with empty balance arrays
//  2. Compute SHA-256 hash of owner's public key
//  3. Extract wallet UUID from created asset
//  4. Create UserDirectory entry mapping public key hash to wallet UUID
//
// Returns:
//   - JSON representation of the created wallet
//   - Error if wallet creation fails or directory entry cannot be saved
//
// Security: The ownerCertHash is required for all subsequent wallet operations,
// ensuring only the legitimate owner can access or modify the wallet.
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
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
		walletId, _ := req["walletId"].(string)
		ownerPublicKey, _ := req["ownerPubKey"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)

		hash := sha256.Sum256([]byte(ownerPublicKey))
		pubKeyHash := hex.EncodeToString(hash[:])

		walletMap := make(map[string]any)
		walletMap["@assetType"] = "wallet"
		walletMap["walletId"] = walletId
		walletMap["ownerPubKey"] = ownerPublicKey
		walletMap["ownerCertHash"] = ownerCertHash
		walletMap["escrowBalances"] = make([]any, 0)
		walletMap["balances"] = make([]any, 0)
		walletMap["digitalAssetTypes"] = make([]any, 0)
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

		userDirMap := make(map[string]any)
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

// GetBalance retrieves the available (non-escrowed) balance for a specific token in a wallet.
// This operation requires certificate-based authentication to prevent unauthorized balance queries.
//
// Arguments:
//   - pubKey: Public key of the wallet owner
//   - assetSymbol: Symbol of the digital asset to query (e.g., "USDT")
//   - ownerCertHash: Certificate hash for ownership verification
//
// Process Flow:
//  1. Compute public key hash and lookup wallet UUID via UserDirectory
//  2. Retrieve wallet from ledger
//  3. Verify owner authorization via certificate hash
//  4. Iterate through wallet's asset types to find matching symbol
//  5. Return corresponding balance from parallel balances array
//
// Returns:
//   - JSON response with wallet ID, asset symbol, and available balance
//   - Error if wallet not found, unauthorized, or asset not held
//
// Note: This returns only the available balance, not the escrowed balance.
var GetBalance = transactions.Transaction{
	Tag:         "getBalance",
	Label:       "Get Wallet Balance",
	Description: "Get balance of a specific token in wallet with authentication",
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
			Tag:      "pubKey",
			Label:    "Public Key",
			DataType: "string",
			Required: true,
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

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
		pubKey, _ := req["pubKey"].(string)
		assetSymbol, _ := req["assetSymbol"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)

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
		walletId := userDir.GetProp("walletUUID").(string)

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
		digitalAssetTypes := walletAsset.GetProp("digitalAssetTypes").([]any)
		balances := walletAsset.GetProp("balances").([]any)

		for i, assetRef := range digitalAssetTypes {
			// Get the referenced asset
			var assetKey string
			switch ref := assetRef.(type) {
			case map[string]any:
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
				response := map[string]any{
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

// GetEscrowBalance retrieves the locked (escrowed) balance for a specific token in a wallet.
// Escrowed tokens are temporarily unavailable for spending while locked in active escrow contracts.
//
// Arguments:
//   - pubKey: Public key of the wallet owner
//   - assetSymbol: Symbol of the digital asset to query
//   - ownerCertHash: Certificate hash for ownership verification
//
// Process Flow:
//  1. Resolve wallet UUID from public key hash
//  2. Retrieve wallet and verify ownership
//  3. Find asset index by matching symbol
//  4. Return corresponding escrow balance
//
// Returns:
//   - JSON response with wallet ID, asset symbol, and escrowed balance
//   - Error if wallet not found, unauthorized, or asset not held
//
// Use Cases:
//   - Verify sufficient funds are locked before attempting escrow release
//   - Display total wallet balance (available + escrowed) in user interfaces
//   - Audit escrow participation for compliance reporting
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
		// {Tag: "walletUUID", DataType: "string", Required: true},
		{Tag: "pubKey", Label: "Public Key", DataType: "string", Required: true},
		{Tag: "assetSymbol", DataType: "string", Required: true},
		{Tag: "ownerCertHash", DataType: "string", Required: true},
	},
	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
		pubKey, _ := req["pubKey"].(string)
		assetSymbol, _ := req["assetSymbol"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)

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
		walletId := userDir.GetProp("walletUUID").(string)

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
		digitalAssetTypes := walletAsset.GetProp("digitalAssetTypes").([]any)
		escrowBalances := walletAsset.GetProp("escrowBalances").([]any)

		for i, assetRef := range digitalAssetTypes {
			// Get the referenced asset
			var assetKey string
			switch ref := assetRef.(type) {
			case map[string]any:
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
				response := map[string]any{
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

// GetWalletByOwner retrieves complete wallet details using the owner's public key.
// This operation returns the full wallet state including all balances, escrow balances,
// and asset types held, subject to ownership verification.
//
// Arguments:
//   - pubKey: Public key of the wallet owner
//   - ownerCertHash: Certificate hash for ownership verification
//
// Process Flow:
//  1. Compute public key hash and lookup wallet UUID via UserDirectory
//  2. Retrieve complete wallet asset from ledger
//  3. Verify owner authorization via certificate hash
//  4. Return full wallet JSON representation
//
// Returns:
//   - JSON representation of the complete wallet state
//   - Error if wallet not found or authorization fails
//
// Security: Certificate verification ensures only the wallet owner can view
// their complete wallet details, maintaining transaction privacy.
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
			Tag:      "pubKey",
			Label:    "Public Key",
			DataType: "string",
			Required: true,
		},
		{
			Tag:         "ownerCertHash",
			Label:       "Owner Certificate Hash",
			Description: "Certificate hash for authentication",
			DataType:    "string",
			Required:    true,
		},
	},

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
		pubKey, _ := req["pubKey"].(string)
		ownerCertHash, _ := req["ownerCertHash"].(string)

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
		walletUuid := userDir.GetProp("walletUUID").(string)

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
