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
			Tag:      "balance",
			Label:    "Token Balance",
			DataType: "number",
			Required: true,
		},
		{
			Tag:      "digitalAssetType",
			Label:    "Asset Type Reference",
			DataType: "string", // References digitalAsset
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
