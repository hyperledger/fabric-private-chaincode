package assettypes

import (
	"fmt"

	"github.com/hyperledger-labs/cc-tools/assets"
)

var Person = assets.AssetType{
	Tag:         "person",
	Label:       "Person",
	Description: "Personal data of someone",

	Props: []assets.AssetProp{
		{
			// Primary key
			Required: true,
			IsKey:    true,
			Tag:      "id",
			Label:    "CPF (Brazilian ID)",
			DataType: "cpf",                          // Datatypes are identified at datatypes folder
			Writers:  []string{`org1MSP`, "Org1MSP"}, // This means only org1 can create the asset (others can edit)
		},
		{
			// Mandatory property
			Required: true,
			Tag:      "name",
			Label:    "Name of the person",
			DataType: "string",
			// Validate funcion
			Validate: func(name interface{}) error {
				nameStr := name.(string)
				if nameStr == "" {
					return fmt.Errorf("name must be non-empty")
				}
				return nil
			},
		},
		{
			// Optional property
			Tag:      "dateOfBirth",
			Label:    "Date of Birth",
			DataType: "datetime",
		},
		{
			// Property with default value
			Tag:          "height",
			Label:        "Person's height",
			DefaultValue: 0,
			DataType:     "number",
		},
	},
}
