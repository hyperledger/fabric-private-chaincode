package assettypes

import "github.com/hyperledger-labs/cc-tools/assets"

// Description of a Library as a collection of books
var Library = assets.AssetType{
	Tag:         "library",
	Label:       "Library",
	Description: "Library as a collection of books",

	Props: []assets.AssetProp{
		{
			// Primary Key
			Required: true,
			IsKey:    true,
			Tag:      "name",
			Label:    "Library Name",
			DataType: "string",
			Writers:  []string{`org3MSP`, "Org1MSP"}, // This means only org3 can create the asset (others can edit)
		},
		{
			// Asset reference list
			Tag:      "books",
			Label:    "Book Collection",
			DataType: "[]->book",
		},
		{
			// Asset reference list
			Tag:      "entranceCode",
			Label:    "Entrance Code for the Library",
			DataType: "->secret",
		},
	},
}
