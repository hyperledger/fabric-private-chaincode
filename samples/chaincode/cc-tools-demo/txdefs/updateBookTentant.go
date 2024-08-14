package txdefs

import (
	"encoding/json"

	"github.com/hyperledger-labs/cc-tools/accesscontrol"
	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/errors"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	tx "github.com/hyperledger-labs/cc-tools/transactions"
)

// Updates the tenant of a Book
// POST Method
var UpdateBookTenant = tx.Transaction{
	Tag:         "updateBookTenant",
	Label:       "Update Book Tenant",
	Description: "Change the tenant of a book",
	Method:      "PUT",
	Callers: []accesscontrol.Caller{ // Any org can call this transaction
		{MSP: `$org\dMSP`},
		{MSP: "Org1MSP"},
	},

	Args: []tx.Argument{
		{
			Tag:         "book",
			Label:       "Book",
			Description: "Book",
			DataType:    "->book",
			Required:    true,
		},
		{
			Tag:         "tenant",
			Label:       "tenant",
			Description: "New tenant of the book",
			DataType:    "->person",
		},
	},
	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		bookKey, ok := req["book"].(assets.Key)
		if !ok {
			return nil, errors.WrapError(nil, "Parameter book must be an asset")
		}
		tenantKey, ok := req["tenant"].(assets.Key)
		if !ok {
			return nil, errors.WrapError(nil, "Parameter tenant must be an asset")
		}

		// Returns Book from channel
		bookAsset, err := bookKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "failed to get asset from the ledger", err.Status())
		}
		bookMap := (map[string]interface{})(*bookAsset)

		// Returns person from channel
		tenantAsset, err := tenantKey.Get(stub)
		if err != nil {
			return nil, errors.WrapErrorWithStatus(err, "failed to get asset from the ledger", err.Status())
		}
		tenantMap := (map[string]interface{})(*tenantAsset)

		updatedTenantKey := make(map[string]interface{})
		updatedTenantKey["@assetType"] = "person"
		updatedTenantKey["@key"] = tenantMap["@key"]

		// Update data
		bookMap["currentTenant"] = updatedTenantKey

		bookMap, err = bookAsset.Update(stub, bookMap)
		if err != nil {
			return nil, errors.WrapError(err, "failed to update asset")
		}

		// Marshal asset back to JSON format
		bookJSON, nerr := json.Marshal(bookMap)
		if nerr != nil {
			return nil, errors.WrapError(err, "failed to marshal response")
		}

		return bookJSON, nil
	},
}
