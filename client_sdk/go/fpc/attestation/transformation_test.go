/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"log"
	"testing"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
)

func Test(t *testing.T) {
	var err error

	credentials := &protos.Credentials{}

	credentials.Attestation = []byte(`{"attestation_type":"simulated","attestation":"MA=="}`)
	credentials, err = ToEvidence(credentials)
	if err != nil {
		log.Fatalf("conversion failed: %v", err)
	}
	expected := `{"attestation_type":"simulated","evidence":"MA=="}`
	if expected != string(credentials.Evidence) {
		log.Fatalf("conversion provided '%v' rather than expected '%v'", string(credentials.Evidence), expected)
	}
}
