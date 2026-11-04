/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// this tool is meant to be used in `$FPC_PATH/common/crypto/attestation-api/test` to ensure compatibility
// with the shell-based attestation conversion implementation in `$FPC_PATH/common/crypto/attestation-api/conversion`.
package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/pkg/errors"
)

func printHelp() {
	fmt.Printf(
		`Usage: %s [<attestation as json>]
convert attestation to evidence in (base64-encoded) Credentials protobuf
`,
		os.Args[0])
}

func main() {

	// get input
	if len(os.Args) < 2 {
		printHelp()
		exitIfError(fmt.Errorf("expect argument"))
	}

	// convert
	output, err := convert([]byte(os.Args[1]))
	exitIfError(err)

	// return output
	fmt.Printf("%s\n", string(output))
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func convert(input []byte) ([]byte, error) {
	credentials := &protos.Credentials{
		SerializedAttestedData: nil,
		Attestation:            input,
		Evidence:               nil,
	}
	credentialsOnlyAttestation := utils.MarshallProtoBase64(credentials)

	// conversion
	converter := attestation.NewDefaultCredentialConverter()
	credentialsStringOut, err := converter.ConvertCredentials(credentialsOnlyAttestation)
	if err != nil {
		return nil, errors.Wrap(err, "ERROR: couldn't convert credentials")
	}

	credentialsOut, err := utils.UnmarshalCredentials(credentialsStringOut)
	if err != nil {
		return nil, errors.Wrap(err, "ERROR: couldn't unmarshal credentials")
	}

	// return to stdout
	return credentialsOut.Evidence, nil
}
