package assets

import (
	"github.com/hyperledger-labs/cc-tools/assets"
)

// Wallet represents a confidential user wallet
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
			Tag:      "ownerId",
			Label:    "Owner Identity",
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
			Required: true,
		},
		{
			Tag:      "digitalAssetTypes",
			Label:    "Asset Type Reference",
			DataType: "[]->digitalAsset", // References digitalAsset
			Required: true,
		},
		{
			Tag:      "createdAt",
			Label:    "Creation Timestamp",
			DataType: "datetime",
			Required: false,
		},
	},
}
