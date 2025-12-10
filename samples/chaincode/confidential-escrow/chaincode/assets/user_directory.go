package assets

import (
	"github.com/hyperledger-labs/cc-tools/assets"
)

// UserDirectory provides a mapping between user public key hashes and wallet UUIDs.
// This indirection layer enables efficient wallet lookups while maintaining privacy,
// as the actual public keys are never stored directly on the ledger.
//
// Purpose:
//   - Enable wallet discovery using only the public key hash
//   - Associate certificate hashes with wallets for authorization
//   - Maintain user privacy by avoiding direct public key storage
//
// Usage Pattern:
//  1. Hash the user's public key with SHA-256
//  2. Query UserDirectory using the hash
//  3. Retrieve the associated wallet UUID
//  4. Access the wallet using the UUID
var UserDirectory = assets.AssetType{
	Tag:         "userdir",
	Label:       "User Directory",
	Description: "Maps user public key hash to wallet ID for authentication",

	Props: []assets.AssetProp{
		{
			Tag:      "publicKeyHash",
			Label:    "Public Key Hash",
			DataType: "string",
			Required: true,
			IsKey:    true,
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
}
