/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// this tool is meant to be used in `$FPC_PATH/common/crypto/attestation-api/test` to ensure compatibility
// with the shell-based attestation verification implementation in `$FPC_PATH/common/crypto/attestation-api/evidence`.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/epid/pdo"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/simulation"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"
)

func main() {

	evidenceJson, err := readFile("verify_evidence_input.txt")
	exitIfError(err)

	statementJson, err := readFile("statement.txt")
	exitIfError(err)

	expectedMrenclave, err := readFile("code_id.txt")
	exitIfError(err)

	verifier := attestation.NewCredentialVerifier(
		simulation.NewSimulationVerifier(),
		pdo.NewEpidLinkableVerifier(),
		pdo.NewEpidUnlinkableVerifier(),
	)

	cred := &protos.Credentials{
		SerializedAttestedData: &anypb.Any{
			Value: []byte(statementJson),
		},
		Evidence: []byte(evidenceJson),
	}

	err = verifier.VerifyCredentials(cred, expectedMrenclave)
	exitIfError(err)
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func readFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "could not read %s", path)
	}

	if len(content) == 0 {
		return "", errors.Errorf("empty file %s", path)
	}

	return strings.TrimSuffix(string(content), "\n"), nil
}
