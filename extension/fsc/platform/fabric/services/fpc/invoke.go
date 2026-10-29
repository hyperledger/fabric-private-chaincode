/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"github.com/pkg/errors"
)

// ChaincodeInvocation models the invocation of an FPC
type ChaincodeInvocation struct {
	*Chaincode

	function string
	args     []interface{}
}

// Call invokes the chaincode and returns the result
func (i *ChaincodeInvocation) Call() ([]byte, error) {
	args, err := argsToStrings(i.args)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed preparing arguments")
	}
	return i.InvokerContract.SubmitTransaction(i.function, args...)
}
