/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package simulation

import (
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"
)

const SimulationType = "simulated"

// NewSimulationConverter creates a new attestation converter for Intel SGX simulation mode
func NewSimulationConverter() *types.Converter {
	return &types.Converter{
		Type: SimulationType,
		Converter: func(attestationBytes []byte) (evidenceBytes []byte, err error) {
			// NO-OP
			return attestationBytes, nil
		},
	}
}
