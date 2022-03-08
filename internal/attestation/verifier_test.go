/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/simulation"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"
	"github.com/stretchr/testify/assert"
)

func NewDummyVerifier() *types.Verifier {
	return &types.Verifier{
		Type: "dummy",
		Verify: func(evidence *types.Evidence, expectedValidationValues *types.ValidationValues) (err error) {
			return nil
		},
	}
}

func TestVerifier(t *testing.T) {

	d := newVerifierDispatcher()

	ev := &types.Evidence{
		Type: "dummy",
	}

	ref := &types.ValidationValues{
		Statement: nil,
		Mrenclave: "",
	}

	// should fail as no converter yet registered for type dummy
	err := d.Verify(ev, ref)
	assert.Error(t, err)

	// register dummy converter
	err = d.Register(NewDummyVerifier())
	assert.NoError(t, err)

	// conversion should now succeed
	err = d.Verify(ev, ref)
	assert.NoError(t, err)

	// trying to register dummy again should fail as already registered
	err = d.Register(NewDummyVerifier())
	assert.Error(t, err)

	// but registering another converter should work fine
	err = d.Register(simulation.NewSimulationVerifier())
	assert.NoError(t, err)
}
