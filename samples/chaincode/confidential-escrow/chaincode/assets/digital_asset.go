package assets

import (
	"github.com/hyperledger-labs/cc-tools/assets"
)

var DigitalAssetToken = assets.AssetType{
	Tag:         "digitalAsset",
	Label:       "Digital Asset Token",
	Description: "Confidential digital currency token (e.g., CBDC)",

	Props: []assets.AssetProp{
		{
			Tag:      "name",
			Label:    "Token Name",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "symbol",
			Label:    "Token Symbol",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "decimals",
			Label:    "Decimal Places",
			DataType: "number",
			Required: true,
		},
		{
			Tag:      "totalSupply",
			Label:    "Total Supply",
			DataType: "number",
			Required: true,
		},
		{
			Tag:      "issuerHash",
			Label:    "Issuer Certificate Hash",
			DataType: "string",
			Required: true,
		},
	},

	Readers: []string{"$org1MSP", "$org2MSP"},
}
