/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package simulation

import (
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"
)

func NewSimulationVerifier() *types.Verifier {
	return &types.Verifier{
		Type: SimulationType,
		Verify: func(evidence *types.Evidence, expectedValidationValues *types.ValidationValues) (err error) {
			// NO-OP
			return nil
		},
	}
}
