/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
)

func ToEvidence(credentials *protos.Credentials) (*protos.Credentials, error) {

	log.Printf("Perform attestation to evidence transformation\n")

	fpcPath := os.Getenv("FPC_PATH")
	if fpcPath == "" {
		return nil, fmt.Errorf("FPC_PATH not set")
	}
	convertScript := filepath.Join(fpcPath, "common/crypto/attestation-api/conversion/attestation_to_evidence.sh")
	cmd := exec.Command(convertScript, string(credentials.Attestation))
	if out, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf("Attestation conversion failed with error %v", err)
	} else {
		credentials.Evidence = []byte(strings.TrimSuffix(string(out), "\n"))
	}
	return credentials, nil
}
