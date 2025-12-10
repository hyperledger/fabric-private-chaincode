package assets

import (
	"github.com/hyperledger-labs/cc-tools/assets"
)

// DigitalAssetToken defines the asset type for fungible digital tokens.
// This represents confidential digital currencies such as Central Bank Digital Currencies (CBDC)
// or tokenized assets. Each token type has a fixed supply controlled by the issuer.
//
// Security: The issuerHash ensures only authorized entities can mint/burn tokens.
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
			IsKey:    true,
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
		{
			Tag:      "owner",
			Label:    "Owner Identity",
			DataType: "string",
			Required: true,
		},
		{
			Tag:      "issuedAt",
			Label:    "Issued At",
			DataType: "datetime",
			Required: false,
		},
	},
}
