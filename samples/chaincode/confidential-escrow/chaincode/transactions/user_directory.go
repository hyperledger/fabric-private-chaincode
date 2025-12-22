// This file implements UserDirectory operations for mapping public key hashes to wallet UUIDs.
// The directory provides a privacy-preserving lookup mechanism enabling wallet discovery
// without exposing actual public keys on the ledger.
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

// CreateUserDir registers a new user directory entry linking a public key hash to a wallet.
// This entry is created automatically during wallet creation but can also be invoked
// independently for manual directory management.
//
// Arguments:
//   - publicKeyHash: SHA-256 hash of the user's public key
//   - walletUUID: UUID of the associated wallet
//   - certHash: Certificate hash of the wallet owner
//
// Returns:
//   - JSON representation of the created directory entry
//   - Error if entry creation or persistence fails
//
// Note: The publicKeyHash serves as the primary key, ensuring one wallet per public key hash.
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

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
		publicKeyHash, _ := req["publicKeyHash"].(string)
		walletId, _ := req["walletUUID"].(string)
		certHash, _ := req["certHash"].(string)

		userDirMap := make(map[string]any)
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

// ReadUserDir retrieves a user directory entry with ownership verification.
// This operation requires the caller to provide a valid certificate hash,
// preventing unauthorized directory lookups.
//
// Arguments:
//   - userDir: Reference to the user directory entry (by publicKeyHash)
//   - certHash: Certificate hash for ownership verification
//
// Returns:
//   - JSON representation of the directory entry including wallet UUID
//   - Error if entry not found or certificate hash mismatch
//
// Security: Certificate verification prevents enumeration attacks where an
// adversary attempts to map all public keys to wallets.
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

	Routine: func(stub *sw.StubWrapper, req map[string]any) ([]byte, errors.ICCError) {
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
