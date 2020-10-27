/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"log"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
)

func ToEvidence(credentials *protos.Credentials) (*protos.Credentials, error) {

	// do something
	log.Printf("Perform attestation to evidence transformation\n")

	// TODO call brunos attestation_to_evidence.sh
	// call $FPC_PATH/common/crypto/attestation-api/conversion/attestation_to_evidence.sh

	//credentials.Evidence = credentials.Attestation

	return credentials, nil
}
