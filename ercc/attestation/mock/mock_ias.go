/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mock

import (
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	"github.com/hyperledger/fabric/common/flogging"
)

type MockIAS struct {
}

var logger = flogging.MustGetLogger("ercc.ias")

func (ias *MockIAS) RequestAttestationReport(apiKey string, quoteAsBytes []byte) (attestation.IASAttestationReport, error) {
	report := attestation.IASAttestationReport{
		IASReportSignature:          "some X-IASReport-Signature",
		IASReportSigningCertificate: "some X-IASReport-Signing-Certificate",
		IASReportBody:               []byte("Some report body"),
	}
	logger.Debugf("Returning empty IAS attestation report (simulation mode)")
	return report, nil
}

func (ias *MockIAS) GetIntelVerificationKey() (interface{}, error) {
	return attestation.PublicKeyFromPem([]byte(attestation.IntelPubPEM))
}
