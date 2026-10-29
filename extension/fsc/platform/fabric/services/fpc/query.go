/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"github.com/pkg/errors"
)

// ChaincodeQuery models the query of an FPC
type ChaincodeQuery struct {
	*Chaincode

	function string
	args     []interface{}
}

// Call invokes the chaincode and returns the result
func (i *ChaincodeQuery) Call() ([]byte, error) {
	logger.Debugf("query FPC [%s:%s:%s]", i.Channel, i.ID, i.function)
	args, err := argsToStrings(i.args)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed preparing arguments")
	}
	return i.InvokerContract.EvaluateTransaction(i.function, args...)
}
