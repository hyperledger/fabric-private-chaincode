/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/contract"
	"github.com/pkg/errors"
)

type transaction struct {
	fns               *fabric.NetworkService
	ch                *fabric.Channel
	id                string
	function          string
	endorserEndpoints []string
	invoker           view.Identity
	txID              fabric.TxID
}

func (t *transaction) Evaluate(args ...string) ([]byte, error) {
	var passedArgs []interface{}
	for _, arg := range args {
		passedArgs = append(passedArgs, arg)
	}
	raw, err := t.ch.Chaincode(t.id).Query(
		t.function, passedArgs...,
	).WithInvokerIdentity(
		t.invoker,
	).WithDiscoveredEndorsersByEndpoints(
		t.endorserEndpoints...,
	).WithTxID(
		t.txID,
	).Call()
	if err != nil {
		return nil, err
	}
	return raw, nil
}

type contract struct {
	fns *fabric.NetworkService
	ch  *fabric.Channel
	id  string
}

func (c *contract) Name() string {
	return c.id
}

func (c *contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	var passedArgs []interface{}
	for _, arg := range args {
		passedArgs = append(passedArgs, arg)
	}
	raw, err := c.ch.Chaincode(c.id).Query(
		name, passedArgs...,
	).WithInvokerIdentity(
		c.fns.IdentityProvider().DefaultIdentity(),
	).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "failed querying chaincode [%s:%s]", c.id, name)
	}
	return raw, err
}

func (c *contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	var passedArgs []interface{}
	for _, arg := range args {
		passedArgs = append(passedArgs, arg)
	}
	_, raw, err := c.ch.Chaincode(c.id).Invoke(
		name, passedArgs...,
	).WithInvokerIdentity(
		c.fns.IdentityProvider().DefaultIdentity(),
	).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "failed invoking chaincode [%s:%s]", c.id, name)
	}
	return raw, err
}

func (c *contract) CreateTransaction(name string, endorsersEndpoints ...string) (fpc.Transaction, error) {
	return &transaction{
		fns:               c.fns,
		ch:                c.ch,
		id:                c.id,
		function:          name,
		endorserEndpoints: endorsersEndpoints,
		invoker:           c.fns.IdentityProvider().DefaultIdentity(),
	}, nil
}

type contractProvider struct {
	fns *fabric.NetworkService
	ch  *fabric.Channel
}

func (c *contractProvider) GetContract(id string) fpc.Contract {
	return &contract{
		fns: c.fns,
		ch:  c.ch,
		id:  id,
	}
}

type contractProviderForEndorsement struct {
	fns     *fabric.NetworkService
	ch      *fabric.Channel
	cid     string
	env     *fabric.Envelope
	invoker view.Identity
	txID    fabric.TxID
}

func (c *contractProviderForEndorsement) GetContract(id string) fpc.Contract {
	if id == "ercc" {
		return &contract{
			fns: c.fns,
			ch:  c.ch,
			id:  id,
		}
	}
	c.cid = id
	return c
}

func (c *contractProviderForEndorsement) Name() string {
	return c.cid
}

func (c *contractProviderForEndorsement) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	var passedArgs []interface{}
	for _, arg := range args {
		passedArgs = append(passedArgs, arg)
	}
	raw, err := c.ch.Chaincode(c.cid).Query(
		name, passedArgs...,
	).WithInvokerIdentity(
		c.invoker,
	).WithTxID(
		c.txID,
	).Call()
	if err != nil {
		return nil, errors.Wrapf(err, "failed querying chaincode [%s:%s]", c.cid, name)
	}
	return raw, nil
}

func (c *contractProviderForEndorsement) SubmitTransaction(name string, args ...string) ([]byte, error) {
	var passedArgs []interface{}
	for _, arg := range args {
		passedArgs = append(passedArgs, arg)
	}
	var err error
	c.env, err = c.ch.Chaincode(c.cid).Endorse(
		name, passedArgs...,
	).WithInvokerIdentity(
		c.invoker,
	).WithTxID(
		c.txID,
	).Call()

	if err != nil {
		return nil, errors.Wrapf(err, "failed invoking chaincode [%s:%s]", c.cid, name)
	}
	return nil, nil
}

func (c *contractProviderForEndorsement) CreateTransaction(name string, endorserEndpoints ...string) (fpc.Transaction, error) {
	return &transaction{
		fns:               c.fns,
		ch:                c.ch,
		id:                c.cid,
		txID:              c.txID,
		function:          name,
		endorserEndpoints: endorserEndpoints,
		invoker:           c.invoker,
	}, nil
}

type endorserContractImpl struct {
	fns     *fabric.NetworkService
	ch      *fabric.Channel
	cid     string
	invoker view.Identity
	txID    fabric.TxID
}

func (e *endorserContractImpl) WithInvokerIdentity(identity view.Identity) {
	e.invoker = identity
}

func (e *endorserContractImpl) WithTxID(id fabric.TxID) {
	e.txID = id
}

func (e *endorserContractImpl) EndorseTransaction(name string, args ...string) (*fabric.Envelope, error) {
	p := &contractProviderForEndorsement{
		fns:     e.fns,
		ch:      e.ch,
		cid:     e.cid,
		invoker: e.invoker,
		txID:    e.txID,
	}
	c := fpc.GetContract(p, e.cid)
	_, err := c.SubmitTransaction(name, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed endorse transaction for [%s:%s]", e.cid, name)
	}
	return p.env, nil
}
