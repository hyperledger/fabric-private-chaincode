package assettypes

import "github.com/hyperledger-labs/cc-tools/assets"

// Secret is and information available only for org2 and org3
// Collections.json configuration is necessary
var Secret = assets.AssetType{
	Tag:         "secret",
	Label:       "Secret",
	Description: "Secret between Org2 and Org3",

	Readers: []string{"org2MSP", "org3MSP", "Org1MSP"},
	Props: []assets.AssetProp{
		{
			// Primary Key
			IsKey:    true,
			Tag:      "secretName",
			Label:    "Secret Name",
			DataType: "string",
			Writers:  []string{`org2MSP`, "Org1MSP"}, // This means only org2 can create the asset (org3 can edit)
		},
		{
			// Mandatory Property
			Required: true,
			Tag:      "secret",
			Label:    "Secret",
			DataType: "string",
		},
	},
}
