/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

type MockVerifier struct {
}

func (v *MockVerifier) VerifyAttestationReport(verificationPubKey interface{}, report IASAttestationReport) (bool, error) {
	return true, nil
}

func (v *MockVerifier) CheckMrEnclave(mrEnclaveBase64 string, report IASAttestationReport) (bool, error) {
	return true, nil
}

func (v *MockVerifier) CheckEnclavePkHash(pkBytes []byte, report IASAttestationReport) (bool, error) {
	return true, nil
}
