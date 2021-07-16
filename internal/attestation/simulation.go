/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

// NewSimulationConverter creates a new attestation converter for Intel SGX simulation mode
func NewSimulationConverter() *Converter {
	return &Converter{
		Type: "simulated",
		Converter: func(attestationBytes []byte) (evidenceBytes []byte, err error) {
			return attestationBytes, nil
		},
	}
}
