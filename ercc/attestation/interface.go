package attestation

type VerifierInterface interface {
	VerifyEvidence(evidenceBytes, expectedStatementBytes []byte, expectedMrEnclave string) error
}
