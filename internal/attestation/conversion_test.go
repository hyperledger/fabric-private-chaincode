/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/simulation"

	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/stretchr/testify/assert"
)

func NewDummyConverter() *types.Converter {
	return &types.Converter{
		Type: "dummy",
		Converter: func(attestationBytes []byte) (evidenceBytes []byte, err error) {
			return []byte("dummy evidence"), nil
		},
	}
}

func TestDispatcher(t *testing.T) {

	d := newConverterDispatcher()

	attestation := &types.Attestation{
		Type: "dummy",
	}

	// should fail as no converter yet registered for type dummy
	evidence, err := d.Convert(attestation)
	assert.Error(t, err)
	assert.Nil(t, evidence)

	// register dummy converter
	err = d.Register(NewDummyConverter())
	assert.NoError(t, err)

	// conversion should now succeed
	evidence, err = d.Convert(attestation)
	assert.NoError(t, err)
	assert.Equal(t, "dummy", evidence.Type)
	assert.Equal(t, "dummy evidence", evidence.Data)

	// trying to register dummy again should fail as already registered
	err = d.Register(NewDummyConverter())
	assert.Error(t, err)

	// but registering another converter should work fine
	err = d.Register(simulation.NewSimulationConverter())
	assert.NoError(t, err)
}

func TestCredentialConverterWithSimulation(t *testing.T) {

	att := []byte(`{"attestation_type":"simulated","attestation":"MA=="}`)
	expectedEvidence := []byte(`{"attestation_type":"simulated","evidence":"MA=="}`)

	credentials := &protos.Credentials{
		Attestation: att,
	}

	cv := NewDefaultCredentialConverter()

	serializedCredentials := utils.MarshallProtoBase64(credentials)
	updatedSerializedCredentials, err := cv.ConvertCredentials(serializedCredentials)
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedSerializedCredentials)

	updatedCredentials, err := utils.UnmarshalCredentials(updatedSerializedCredentials)
	assert.NoError(t, err)
	assert.NotNil(t, updatedCredentials)

	assert.Equal(t, att, updatedCredentials.Attestation)
	assert.Equal(t, expectedEvidence, updatedCredentials.Evidence)
}
