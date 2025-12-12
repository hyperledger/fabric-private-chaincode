package assets

import (
	"github.com/hyperledger-labs/cc-tools/assets"
)

// Escrow defines the asset type for programmable conditional payment contracts.
// This enables secure, trustless transactions where funds are held in escrow until
// predefined conditions are met. The escrow uses cryptographic hash verification
// to ensure condition fulfillment.
//
// Lifecycle States:
//   - Active: Funds locked, awaiting condition verification
//   - ReadyForRelease: Condition verified, awaiting release
//   - Released: Funds transferred to seller
//   - Refunded: Funds returned to buyer
//
// Security Model:
//   - conditionValue: SHA-256 hash of (secret + parcelId) for atomic condition verification
//   - buyerCertHash: Ensures only the buyer can initiate refunds
//   - Seller must provide correct secret and parcelId to release funds
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
