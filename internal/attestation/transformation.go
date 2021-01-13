/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("fpc-client-attest")

func toEvidence(credentials *protos.Credentials) (*protos.Credentials, error) {

	logger.Debugf("Perform attestation to evidence transformation")

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

// perform attestation evidence transformation
func ConvertCredentials(credentialsOnlyAttestation string) (credentialsWithEvidence string, err error) {
	logger.Debugf("Received Credential: '%s'", credentialsOnlyAttestation)
	credentials, err := utils.UnmarshalCredentials(credentialsOnlyAttestation)
	if err != nil {
		return "", fmt.Errorf("cannot decode credentials: %s", err)
	}

	credentials, err = toEvidence(credentials)
	if err != nil {
		return "", err
	}
	credentialsOnlyAttestation = utils.MarshallProto(credentials)
	logger.Debugf("Converted to Credential: '%s'", credentialsOnlyAttestation)
	return credentialsOnlyAttestation, nil
}
