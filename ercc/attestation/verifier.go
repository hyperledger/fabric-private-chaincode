/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"
	"github.com/pkg/errors"
)

var registry verifierRegistry

type verifierRegistry struct {
	verifiers []*types.Verifier
}

func (vr *verifierRegistry) add(verifier *types.Verifier) {
	for _, v := range vr.verifiers {
		if v.Type == verifier.Type {
			// this type of verifier is already registered
			panic(errors.Errorf("credential verifier of type '%v' already registered!", verifier.Type))
		}
	}

	vr.verifiers = append(vr.verifiers, verifier)
}

func GetAvailableVerifier() *attestation.CredentialVerifier {
	return attestation.NewCredentialVerifier(registry.verifiers...)
}
