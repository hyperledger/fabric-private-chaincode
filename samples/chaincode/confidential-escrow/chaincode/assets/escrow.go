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
			Tag:      "amount",
			Label:    "Escrowed Amount",
			DataType: "number",
			Required: true,
		},
		{
			Tag:      "assetType",
			Label:    "Asset Type Reference",
			DataType: "string", // References digitalAsset symbol
			Required: true,
		},
		{
			Tag:      "conditionType",
			Label:    "Condition Type",
			DataType: "string", // "hashlock", "signature", "timelock"
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

	Readers: []string{"$org1MSP", "$org2MSP"},
}
