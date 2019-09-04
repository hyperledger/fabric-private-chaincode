/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package crypto

// MockVerifier implements Verifier interface!
type MockVerifier struct {
}

// Verify returns true if signature validation of enclave return is correct; other false
func (v *MockVerifier) Verify(txType, encoded_args, responseData []byte, readset, writeset [][]byte, signature, enclavePk []byte) (bool, error) {
	return true, nil
}
