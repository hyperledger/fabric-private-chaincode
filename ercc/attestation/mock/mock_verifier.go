/*
* Copyright IBM Corp. All Rights Reserved.
*
* SPDX-License-Identifier: Apache-2.0
 */

package mock

import "github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"

type MockVerifier struct {
}

func (v *MockVerifier) VerifyAttestationReport(verificationPubKey interface{}, report attestation.IASAttestationReport) (bool, error) {
	return true, nil
}

func (v *MockVerifier) CheckMrEnclave(mrEnclaveBase64 string, report attestation.IASAttestationReport) (bool, error) {
	return true, nil
}

func (v *MockVerifier) CheckEnclavePkHash(pkBytes []byte, report attestation.IASAttestationReport) (bool, error) {
	return true, nil
}
