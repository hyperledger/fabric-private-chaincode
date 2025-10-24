package assets

import (
	"github.com/hyperledger-labs/cc-tools/assets"
)

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
