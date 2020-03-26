/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

type MockIAS struct {
}

func (ias *MockIAS) RequestAttestationReport(apiKey string, quoteAsBytes []byte) (IASAttestationReport, error) {
	report := IASAttestationReport{
		IASReportSignature:          "some X-IASReport-Signature",
		IASReportSigningCertificate: "some X-IASReport-Signing-Certificate",
		IASReportBody:               []byte("Some report body"),
	}
	logger.Debugf("Returning empty IAS attestation report (simulation mode)")
	return report, nil
}

func (ias *MockIAS) GetIntelVerificationKey() (interface{}, error) {
	return PublicKeyFromPem([]byte(IntelPubPEM))
}
