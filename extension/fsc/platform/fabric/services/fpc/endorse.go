/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"github.com/pkg/errors"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
)

// ChaincodeEndorse models the endorsement of an FPC
type ChaincodeEndorse struct {
	*Chaincode

	function string
	args     []interface{}
}

func (i *ChaincodeEndorse) WithTxID(txID fabric.TxID) *ChaincodeEndorse {
	i.EndorserContract.WithTxID(txID)
	return i
}

func (i *ChaincodeEndorse) WithSignerIdentity(identity view.Identity) *ChaincodeEndorse {
	i.EndorserContract.WithInvokerIdentity(identity)
	return i
}

// Call invokes the chaincode and returns a Fabric envelope
func (i *ChaincodeEndorse) Call() (*fabric.Envelope, error) {
	args, err := argsToStrings(i.args)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed preparing arguments")
	}
	return i.EndorserContract.EndorseTransaction(i.function, args...)
}
