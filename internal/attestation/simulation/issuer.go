/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package simulation

import "github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"

// NewSimulationIssuer creates a new attestation issuer for Intel SGX simulation mode
func NewSimulationIssuer() *types.Issuer {
	return &types.Issuer{
		Type:  SimulationType,
		Issue: issue,
	}
}

func issue(customData []byte) ([]byte, error) {
	return []byte("{\"attestation_type\":\"" + SimulationType + "\",\"attestation\":\"MA==\"}"), nil
}
