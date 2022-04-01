//go:build WITH_PDO_CRYPTO
// +build WITH_PDO_CRYPTO

/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2021 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package crypto

func init() {

	// add another test case using the PDO crypto lib
	allTestCases = append(allTestCases,
		testCases{"PDO Crypto", NewPdoCrypto()},
	)
}
