package txdefs

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger-labs/cc-tools/accesscontrol"
	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/errors"
	"github.com/hyperledger-labs/cc-tools/events"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	tx "github.com/hyperledger-labs/cc-tools/transactions"
)

// Create a new Library on channel
// POST Method
var CreateNewLibrary = tx.Transaction{
	Tag:         "createNewLibrary",
	Label:       "Create New Library",
	Description: "Create a New Library",
	Method:      "POST",
	Callers: []accesscontrol.Caller{ // Only org3 admin can call this transaction
		{
			MSP: "org3MSP",
			OU:  "admin",
		},
		{
			MSP: "Org1MSP",
			OU:  "admin",
		},
	},

	Args: []tx.Argument{
		{
			Tag:         "name",
			Label:       "Name",
			Description: "Name of the library",
			DataType:    "string",
			Required:    true,
		},
	},
	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		name, _ := req["name"].(string)

		libraryMap := make(map[string]interface{})
		libraryMap["@assetType"] = "library"
		libraryMap["name"] = name

		libraryAsset, err := assets.NewAsset(libraryMap)
		if err != nil {
			return nil, errors.WrapError(err, "Failed to create a new asset")
		}

		// Save the new library on channel
		_, err = libraryAsset.PutNew(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "Error saving asset on blockchain", err.Status())
		}

		// Marshal asset back to JSON format
		libraryJSON, nerr := json.Marshal(libraryAsset)
		if nerr != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		// Marshall message to be logged
		logMsg, ok := json.Marshal(fmt.Sprintf("New library name: %s", name))
		if ok != nil {
			return nil, errors.WrapError(nil, "failed to encode asset to JSON format")
		}

		// Call event to log the message
		events.CallEvent(stub, "createLibraryLog", logMsg)

		return libraryJSON, nil
	},
}
