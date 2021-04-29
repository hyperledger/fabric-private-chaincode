/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sgx_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/stretchr/testify/assert"
)

func TestAttestationParamsToBase64EncodedJSON(t *testing.T) {
	attestationParams := &sgx.AttestationParams{
		AttestationType: "SomeType",
		HexSpid:         "SomeSpid",
		SigRL:           "SomeSigRL",
	}

	jsonBytes, err := attestationParams.ToBase64EncodedJSON()
	assert.NotNil(t, jsonBytes)
	assert.NoError(t, err)
}

func TestAttestationParamsValidate(t *testing.T) {
	// TODO implement with validate
}

func TestCreateAttestationParamsFromEnvironment(t *testing.T) {
	// should fail
	os.Setenv(sgx.SGXModeEnvKey, "INVALID_VALUE")
	attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
	assert.Nil(t, attestationParams)
	assert.Error(t, err)

	// should fail, no credentials path
	os.Setenv(sgx.SGXModeEnvKey, sgx.SGXModeHwType)
	os.Setenv(sgx.SGXCredentialsPathKey, "")
	attestationParams, err = sgx.CreateAttestationParamsFromEnvironment()
	assert.Nil(t, attestationParams)
	assert.Error(t, err)

	// should fail, credentials path don't have any credentials
	os.Setenv(sgx.SGXCredentialsPathKey, "some/path/to/credentials")
	attestationParams, err = sgx.CreateAttestationParamsFromEnvironment()
	assert.Nil(t, attestationParams)
	assert.Error(t, err)

	// success
	os.Setenv(sgx.SGXModeEnvKey, sgx.SGXModeSimType)
	attestationParams, err = sgx.CreateAttestationParamsFromEnvironment()
	assert.NotNil(t, attestationParams)
	assert.NoError(t, err)
	assert.Equal(t, attestationParams.AttestationType, "simulated")
}

func TestCreateAttestationParamsFromCredentialsPath(t *testing.T) {
	// success
	testPath, err := ioutil.TempDir("/tmp/", "attestation")
	assert.NoError(t, err)
	defer os.RemoveAll(testPath)

	// no spidType, no spid available
	attestationParams, err := sgx.CreateAttestationParamsFromCredentialsPath(testPath)
	assert.Nil(t, attestationParams)
	assert.Error(t, err)

	err = ioutil.WriteFile(filepath.Join(testPath, "spid_type.txt"), []byte("simulation"), 0644)
	assert.NoError(t, err)

	// no spid available
	attestationParams, err = sgx.CreateAttestationParamsFromCredentialsPath(testPath)
	assert.Nil(t, attestationParams)
	assert.Error(t, err)

	err = ioutil.WriteFile(filepath.Join(testPath, "spid.txt"), []byte("EEEEAAAABBBBAAAEEEEAAAABBBBAAAA"), 0644)
	assert.NoError(t, err)

	// TODO once ReadSigRL is implemented
	// write recovaction list to file

	attestationParams, err = sgx.CreateAttestationParamsFromCredentialsPath(testPath)
	assert.NoError(t, err)

	assert.Equal(t, attestationParams.AttestationType, "simulation")
	assert.Equal(t, attestationParams.HexSpid, "EEEEAAAABBBBAAAEEEEAAAABBBBAAAA")
	// TODO once ReadSigRL is implemented
	assert.Equal(t, attestationParams.SigRL, "")
}

func TestReadSPIDType(t *testing.T) {
	testPath, err := ioutil.TempDir("/tmp/", "attestation")
	assert.NoError(t, err)
	defer os.RemoveAll(testPath)

	// does not exists
	spidType, err := sgx.ReadSPIDType(testPath)
	assert.Empty(t, spidType)
	assert.Error(t, err)

	// empty spid_type file
	err = ioutil.WriteFile(filepath.Join(testPath, "spid_type.txt"), nil, 0644)
	assert.NoError(t, err)
	spidType, err = sgx.ReadSPIDType(testPath)
	assert.Empty(t, spidType)
	assert.Error(t, err)

	// success
	err = ioutil.WriteFile(filepath.Join(testPath, "spid_type.txt"), []byte("simulation"), 0644)
	assert.NoError(t, err)
	spidType, err = sgx.ReadSPIDType(testPath)
	assert.Equal(t, spidType, "simulation")
	assert.NoError(t, err)
}

func TestReadSPID(t *testing.T) {
	testPath, err := ioutil.TempDir("/tmp/", "attestation")
	assert.NoError(t, err)
	defer os.RemoveAll(testPath)

	// does not exists
	spid, err := sgx.ReadSPID(testPath)
	assert.Empty(t, spid)
	assert.Error(t, err)

	// empty spid file
	err = ioutil.WriteFile(filepath.Join(testPath, "spid.txt"), nil, 0644)
	assert.NoError(t, err)
	spid, err = sgx.ReadSPID(testPath)
	assert.Empty(t, spid)
	assert.Error(t, err)

	// success
	err = ioutil.WriteFile(filepath.Join(testPath, "spid.txt"), []byte("EEEEAAAABBBBAAAEEEEAAAABBBBAAAA"), 0644)
	assert.NoError(t, err)
	spid, err = sgx.ReadSPID(testPath)
	assert.Equal(t, spid, "EEEEAAAABBBBAAAEEEEAAAABBBBAAAA")
	assert.NoError(t, err)
}

func TestReadSigRL(t *testing.T) {
	// TODO once ReadSigRL is implemented
	sigRL, err := sgx.ReadSigRL("TODO")
	assert.Empty(t, sigRL)
	assert.NoError(t, err)
}
