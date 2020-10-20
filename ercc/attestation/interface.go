/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

type VerifierInterface interface {
	VerifyEvidence(evidenceBytes, expectedStatementBytes []byte, expectedMrEnclave string) error
}
