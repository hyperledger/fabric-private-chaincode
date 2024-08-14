package txdefs

import (
	"encoding/json"

	"github.com/hyperledger-labs/cc-tools/accesscontrol"
	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/errors"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	tx "github.com/hyperledger-labs/cc-tools/transactions"
)

// Return the number of books of a library
// GET method
var GetNumberOfBooksFromLibrary = tx.Transaction{
	Tag:         "getNumberOfBooksFromLibrary",
	Label:       "Get Number Of Books From Library",
	Description: "Return the number of books of a library",
	Method:      "GET",
	Callers: []accesscontrol.Caller{ // Only org2 can call this transaction
		{MSP: "org2MSP"},
		{MSP: "Org1MSP"},
	},

	Args: []tx.Argument{
		{
			Tag:         "library",
			Label:       "Library",
			Description: "Library",
			DataType:    "->library",
			Required:    true,
		},
	},
	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		libraryKey, _ := req["library"].(assets.Key)

		// Returns Library from channel
		libraryMap, err := libraryKey.GetMap(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "failed to get asset from the ledger", err.Status())
		}

		numberOfBooks := 0
		books, ok := libraryMap["books"].([]interface{})
		if ok {
			numberOfBooks = len(books)
		}

		returnMap := make(map[string]interface{})
		returnMap["numberOfBooks"] = numberOfBooks

		// Marshal asset back to JSON format
		returnJSON, nerr := json.Marshal(returnMap)
		if nerr != nil {
			return nil, errors.WrapError(err, "failed to marshal response")
		}

		return returnJSON, nil
	},
}
