/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import "github.com/hyperledger/fabric-private-chaincode/internal/attestation/simulation"

func init() {
	registry.add(simulation.NewSimulationVerifier())
}
