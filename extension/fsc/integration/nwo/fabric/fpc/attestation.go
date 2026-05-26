/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
)

func CreateSIMAttestationParams() *sgx.AttestationParams {
	return &sgx.AttestationParams{
		AttestationType: "simulated",
	}
}
