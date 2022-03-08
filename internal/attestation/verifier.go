/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/pkg/errors"
)

type verifierDispatcher struct {
	verifiers map[string]types.VerifyFunction
}

func newVerifierDispatcher() *verifierDispatcher {
	return &verifierDispatcher{
		verifiers: make(map[string]types.VerifyFunction),
	}
}

// Register adds new verifiers to the verifierDispatcher
func (d *verifierDispatcher) Register(verifiers ...*types.Verifier) error {
	for _, v := range verifiers {
		if _, ok := d.verifiers[v.Type]; ok {
			return fmt.Errorf("'%s' type is already registered", v.Type)
		}
		logger.Debugf("Register verifier of type '%s'", v.Type)
		d.verifiers[v.Type] = v.Verify
	}
	return nil
}

func (d *verifierDispatcher) Verify(evidence *types.Evidence, expectedValidationValues *types.ValidationValues) error {
	verify, ok := d.verifiers[evidence.Type]
	if !ok {
		return fmt.Errorf("'%s' type is not registered", evidence.Type)
	}

	logger.Debugf("Invoke verifier of type '%s'", evidence.Type)
	return verify(evidence, expectedValidationValues)
}

type CredentialVerifier struct {
	dispatcher *verifierDispatcher
}

type Verifier interface {
	VerifyCredentials(credentials *protos.Credentials, expectedMrenclave string) (err error)
}

func NewCredentialVerifier(verifier ...*types.Verifier) *CredentialVerifier {
	dispatcher := newVerifierDispatcher()
	err := dispatcher.Register(verifier...)
	if err != nil {
		// ouch this should never happen
		logger.Panicf("cannot create new credential converter! Reason: %s", err.Error())
	}

	return &CredentialVerifier{dispatcher: dispatcher}
}

func (c *CredentialVerifier) VerifyCredentials(credentials *protos.Credentials, expectedMrenclave string) error {

	evidence, err := unmarshalEvidence(credentials.Evidence)
	if err != nil {
		return err
	}

	expectedValues := &types.ValidationValues{
		Statement: credentials.SerializedAttestedData.Value,
		Mrenclave: expectedMrenclave,
	}

	return c.dispatcher.Verify(evidence, expectedValues)
}

func unmarshalEvidence(serializedEvidence []byte) (*types.Evidence, error) {
	att := &types.Evidence{}
	err := json.Unmarshal(serializedEvidence, att)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal evidence json")
	}

	return att, nil
}
