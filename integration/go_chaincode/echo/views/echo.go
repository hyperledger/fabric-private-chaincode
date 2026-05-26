/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package views

import (
	"encoding/json"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/utils/assert"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger/fabric-private-chaincode/extension/fsc/platform/fabric/services/fpc"
)

// Echo models the parameters to be used to invoke the Echo FPC
type Echo struct {
	// Function to invoke
	Function string
	// Args to pass to the function
	Args []string
}

// EchoView models a View that invokes the Echo FPC
type EchoView struct {
	*Echo
}

func (e *EchoView) Call(context view.Context) (interface{}, error) {
	ch := fpc.GetDefaultChannel(context)
	v, err := ch.EnclaveRegistry().IsAvailable()
	assert.NoError(err, "failed checking availability of the enclave registry")
	assert.True(v, "the enclave registry is not available")

	v, err = ch.EnclaveRegistry().IsPrivate("echo")
	assert.NoError(err, "failed checking echo deployment")
	assert.True(v, "echo should be an FPC")

	v, err = ch.EnclaveRegistry().IsPrivate("mycc")
	assert.NoError(err, "failed checking mycc deployment")
	assert.False(v, "mycc should be a standard CC")

	// Invoke the `echo` chaincode via the FPC-encrypted channel API.
	// Note: do NOT use chaincode.NewInvokeView / NewQueryView / NewEndorseView for FPC
	// chaincodes. Those FSC helpers send plain-text args; the FPC private chaincode
	// wrapper only handles __invoke/__initEnclave/__endorse and rejects all other
	// calls with "invalid invocation". Always use ch.Chaincode(...).Invoke/Query/Endorse.
	res, err := ch.Chaincode(
		"echo",
	).Invoke(
		e.Function, fpc.StringsToArgs(e.Args)...,
	).Call()
	assert.NoError(err, "failed invoking echo")
	assert.Equal(e.Function, string(res))

	// Query the `echo` chaincode via FPC channel
	res, err = ch.Chaincode(
		"echo",
	).Query(
		e.Function, fpc.StringsToArgs(e.Args)...,
	).Call()
	assert.NoError(err, "failed querying echo")
	assert.Equal(e.Function, string(res))

	// Endorse + broadcast via FPC channel
	fns, err := fabric.GetDefaultFNS(context)
	assert.NoError(err)
	envelope, err := ch.Chaincode(
		"echo",
	).Endorse(
		e.Function, fpc.StringsToArgs(e.Args)...,
	).WithSignerIdentity(
		fns.LocalMembership().DefaultIdentity(),
	).Call()
	assert.NoError(err, "failed endorsing echo")
	assert.NotNil(envelope)
	assert.NoError(fns.Ordering().Broadcast(context.Context(), envelope))

	return res, nil
}

type EchoViewFactory struct{}

func (l *EchoViewFactory) NewView(in []byte) (view.View, error) {
	f := &EchoView{}
	assert.NoError(json.Unmarshal(in, &f.Echo))
	return f, nil
}
