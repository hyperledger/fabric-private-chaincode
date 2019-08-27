/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package crypto

// Verifier interface
type Verifier interface {
	Verify(txType, encoded_args, responseData []byte, readset, writeset [][]byte, signature, enclavePk []byte) (bool, error)
}
