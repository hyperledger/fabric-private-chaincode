package transactions

import (
	"github.com/hyperledger-labs/cc-tools/errors"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	"github.com/hyperledger-labs/cc-tools/transactions"
)

var DebugTest = transactions.Transaction{
	Tag:         "debugTest",
	Label:       "Debug Test",
	Description: "Test transaction with no access control",
	Method:      "GET",
	// NO Callers field at all
	Args: []transactions.Argument{},
	Routine: func(stub *sw.StubWrapper, req map[string]interface{}) ([]byte, errors.ICCError) {
		return []byte("Debug test successful"), nil
	},
}
