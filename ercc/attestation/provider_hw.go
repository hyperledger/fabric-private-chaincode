//+build sgx_hw_mode

/*
Copyright 2020 Intel Corporation
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

func GetVerifier() Verifier {
	return &VerifierImpl{}
}
func GetIAS() IntelAttestationService {
	return NewIAS()
}
