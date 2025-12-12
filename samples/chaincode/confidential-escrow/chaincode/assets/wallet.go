package assets

import (
	"github.com/hyperledger-labs/cc-tools/assets"
)

// Wallet represents a confidential user account for holding multiple digital assets.
// Each wallet maintains separate tracking of available balances and escrowed balances
// to ensure accurate accounting during conditional payment operations.
//
// Balance Management:
//   - balances: Freely spendable token amounts
//   - escrowBalances: Tokens locked in active escrow contracts
//   - digitalAssetTypes: References to the types of tokens held
//
// All three arrays are parallel (same length, matching indices) to maintain
// consistency between asset types and their corresponding balances.
//
// Security:
//   - ownerCertHash: Required for all operations to verify wallet ownership
//   - ownerPubKey: Public key associated with the wallet for external verification
var Wallet = assets.AssetType{
	Tag:         "wallet",
	Label:       "User Wallet",
	Description: "Confidential wallet holding digital assets",

	Props: []assets.AssetProp{
		{
			Tag:      "walletId",
			Label:    "Wallet ID",
			DataType: "string",
			Required: true,
			IsKey:    true, // primary key
		},
		{
			Tag:      "ownerPubKey",
			Label:    "Owner Public Key",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "ownerCertHash",
			Label:    "Owner Certificate Hash",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "balances",
			Label:    "Token Balance",
			DataType: "[]number",
			Required: false,
		},
		{
			Tag:      "escrowBalances",
			Label:    "Escrowed Token Balances",
			DataType: "[]number",
			Required: false, // Initialize as empty for existing wallets
		},
		{
			Tag:      "digitalAssetTypes",
			Label:    "Asset Type Reference",
			DataType: "[]->digitalAsset", // References digitalAsset
			Required: false,
		},
		{
			Tag:      "createdAt",
			Label:    "Creation Timestamp",
			DataType: "datetime",
			Required: false,
		},
	},
}
