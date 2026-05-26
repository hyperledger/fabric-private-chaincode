/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
)

type invokerContract interface {
	Name() string
	EvaluateTransaction(name string, args ...string) ([]byte, error)
	SubmitTransaction(name string, args ...string) ([]byte, error)
}

type endorserContract interface {
	EndorseTransaction(name string, args ...string) (*fabric.Envelope, error)
	WithInvokerIdentity(identity view.Identity)
	WithTxID(id fabric.TxID)
}

type identityProvider interface {
	Identity(label string) (view.Identity, error)
}

// Chaincode models a Fabric Private Chaincode
type Chaincode struct {
	Channel          *fabric.Channel
	EnclaveRegistry  *EnclaveRegistry
	InvokerContract  invokerContract
	EndorserContract endorserContract
	Signer           view.Identity
	IdentityProvider identityProvider

	ID string
}

// NewChaincode returns a new chaincode instance
func NewChaincode(
	ch *fabric.Channel,
	er *EnclaveRegistry,
	invoker invokerContract,
	endorser endorserContract,
	id view.Identity,
	ip identityProvider,
	cid string,
) *Chaincode {
	return &Chaincode{
		Channel:          ch,
		EnclaveRegistry:  er,
		InvokerContract:  invoker,
		EndorserContract: endorser,
		Signer:           id,
		IdentityProvider: ip,
		ID:               cid,
	}
}

func (c *Chaincode) IsPrivate() bool {
	return true
}

// Invoke returns an object that models an FPC invocation for the passed function and arguments
func (c *Chaincode) Invoke(function string, args ...interface{}) *ChaincodeInvocation {
	return &ChaincodeInvocation{
		Chaincode: c,
		function:  function,
		args:      args,
	}
}

func (c *Chaincode) Endorse(function string, args ...interface{}) *ChaincodeEndorse {
	return &ChaincodeEndorse{
		Chaincode: c,
		function:  function,
		args:      args,
	}
}

func (c *Chaincode) Query(function string, args ...interface{}) *ChaincodeQuery {
	return &ChaincodeQuery{
		Chaincode: c,
		function:  function,
		args:      args,
	}
}

// StringsToArgs converts a slice of strings into a slace of interface{}
func StringsToArgs(args []string) []interface{} {
	var res []interface{}
	for _, arg := range args {
		res = append(res, arg)
	}
	return res
}
