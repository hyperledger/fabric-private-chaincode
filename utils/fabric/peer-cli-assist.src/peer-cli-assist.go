/*
* Copyright 2019 Intel Corporation
*
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/fpc"
)

func printHelp() {
	fmt.Printf(
		`Usage: %s [attestation2Evidence | createEncryptRequest | processEncryptedResponse]
- attestation2Evidence: convert attestation to evidence in (base64-encoded) Credentials protobuf
- createEncryptRequest: create a (base64-encoded) encrypted fpc chaincode request protobuf
- processEncryptedResponse: decrypt and validate an (base64-encoded) encrypted fpc chaincode response protobuf
Input and outpus are via stdin and stdout, respectively.`,
		os.Args[0])
}

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "ERROR: expected a subcommand\n")
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "attestation2Evidence":
		credentialsIn, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: couldn't read stdin: %v\n", err)
			os.Exit(1)
		}
		credentialsStringOut, err := fpc.ConvertCredentials(string(credentialsIn))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: couldn't convert credentials: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", credentialsStringOut)
	case "createEncryptRequest":
		fmt.Fprintf(os.Stderr, "FATAL: command %s not yet implemented\n", os.Args[1])
		os.Exit(1)
	case "processEncryptedResponse":
		fmt.Fprintf(os.Stderr, "FATAL: command %s not yet implemented\n", os.Args[1])
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "ERROR: Illegal command '%s'\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}

	os.Exit(0)
}
