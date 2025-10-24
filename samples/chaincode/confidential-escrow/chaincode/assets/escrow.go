package assets

import (
	"github.com/hyperledger-labs/cc-tools/assets"
)

var Escrow = assets.AssetType{
	Tag:         "escrow",
	Label:       "Programmable Escrow",
	Description: "Confidential escrow contract with programmable conditions",

	Props: []assets.AssetProp{
		{
			Tag:      "escrowId",
			Label:    "Escrow ID",
			DataType: "string",
			Required: true,
			IsKey:    true,
		},
		{
			Tag:      "buyerPubKey",
			Label:    "Buyer Public Key",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "sellerPubKey",
			Label:    "Seller Public Key",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "buyerWalletUUID",
			Label:    "Buyer Wallet UUID",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "sellerWalletUUID",
			Label:    "Seller Wallet UUID",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "amount",
			Label:    "Escrowed Amount",
			DataType: "number",
			Required: true,
		},
		{
			Tag:      "assetType",
			Label:    "Asset Type Reference",
			DataType: "->digitalAsset", // References digitalAsset symbol
			Required: true,
		},
		{
			Tag:      "parcelId",
			Label:    "Parcel ID",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "conditionValue",
			Label:    "Condition Value",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "status",
			Label:    "Escrow Status",
			DataType: "string", // "Active", "Released", "Refunded"
			Required: true,
		},
		{
			Tag:      "createdAt",
			Label:    "Creation Timestamp",
			DataType: "datetime",
			Required: false,
		},
		{
			Tag:      "buyerCertHash",
			Label:    "Buyer Certificate Hash",
			DataType: "string",
			Required: true,
		},
	},
}
