/*
* Copyright 2019 Intel Corporation
*
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation"
	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("peer-cli-assist")

func printHelp() {
	fmt.Printf(
		`Usage: %s [attestation2Evidence | handleRequestAndResponse <cid> <pipe>]
- attestation2Evidence: convert attestation to evidence in (base64-encoded) Credentials protobuf
  (Input and outpus are via stdin and stdout, respectively.)
- handleRequestAndResponse: handles the encryption of invocation requests as well as the decryption
  of the corresponding responses.
  Expects three parameters
  - <c_ek> the chaincode encryption key, as returned from ercc.QueryChaincodeEncryptionKey 
    (i.e., a base64-encoded string)
  - <pipe> a path to an (existing) fifo file through which the results are communicated back
  As input, expects two lines
  - a (single line!) json string in peer cli format '{"Function": "...", "Args": [...]}' with the invocation params, 
    after which it returns (as single line) the (base64-encoded) ChaincodeRequestMessage protobuf, and then
  - a (base64-encoded) ChaincodeResponseMessage protobuf, after which it will decrypt it and 
    return a json-encoded fabric response protobuf object
`,
		os.Args[0])
	// TODO: above we have to fix the response payload format (json?)
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

		converter := attestation.NewDefaultCredentialConverter()
		credentialsStringOut, err := converter.ConvertCredentials(string(credentialsIn))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: couldn't convert credentials: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", credentialsStringOut)
	case "handleRequestAndResponse":
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "ERROR: command 'handleRequestAndResponse' needs exactly two arguments\n")
			printHelp()
			os.Exit(1)
		}
		handleEncryptedRequestAndResponse(os.Args[2], os.Args[3])
	default:
		fmt.Fprintf(os.Stderr, "ERROR: Illegal command '%s'\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}

	os.Exit(0)
}

func handleEncryptedRequestAndResponse(chaincodeEncryptionKey string, resultPipeName string) {
	reader := bufio.NewReader(os.Stdin)
	resultPipeFile, err := os.OpenFile(resultPipeName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: couldn't open pipe '%s': %v\n", resultPipeName, err)
		os.Exit(1)
	}

	// read request
	requestJSON, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: couldn't read json request: %v\n", err)
		os.Exit(1)
	}
	requestJSON = strings.TrimSpace(requestJSON)
	type Request struct {
		Function *string   `json:"function,omitempty"`
		Args     *[]string `json:"args,omitempty"`
	}
	clearRequest := &Request{}
	dec := json.NewDecoder(bytes.NewReader([]byte(requestJSON)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&clearRequest); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unexpected json '%s': %v\n", requestJSON, err)
		os.Exit(1)
	}
	// Note fabric has two invocation formats, i.e., missing Function means function is Args[0]
	if clearRequest.Args == nil {
		clearRequest.Args = &([]string{})
	}
	if clearRequest.Function == nil {
		if len(*clearRequest.Args) > 0 {
			clearRequest.Function = &((*clearRequest.Args)[0])
			remainingArgs := (*clearRequest.Args)[1:]
			clearRequest.Args = &remainingArgs
		} else {
			emptyString := ""
			clearRequest.Function = &emptyString
		}
	}
	logger.Debugf("Normalized json args '%s' to function='%s'/args='%v'", requestJSON, *clearRequest.Function, *clearRequest.Args)

	// setup crypto context
	ep := &crypto.EncryptionProviderImpl{
		CSP: crypto.GetDefaultCSP(),
		GetCcEncryptionKey: func() ([]byte, error) {
			// TODO: might have to do some re-formatting, e.g., de-hex, here?
			return []byte(chaincodeEncryptionKey), nil
		}}

	ctx, err := ep.NewEncryptionContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not setup crypto context: %v\n", err)
		os.Exit(1)
	}
	logger.Debugf("Setup crypto context based on CC-ek '%v'", chaincodeEncryptionKey)

	// encrypt request ...
	encryptedRequest, err := ctx.Conceal(*clearRequest.Function, *clearRequest.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not encrypt request: %v\n", err)
		os.Exit(1)
	}
	// ... and return it
	logger.Debugf("Transformed request '%v' to '%v' and write to pipe '%s'", clearRequest, encryptedRequest, resultPipeName)
	resultPipeFile.WriteString(fmt.Sprintf("%s\n", encryptedRequest))

	// read encrypted response ...
	encryptedResponse, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: couldn't read encrypted response: %v\n", err)
		os.Exit(1)
	}
	encryptedResponse = strings.TrimSuffix(encryptedResponse, "\n")

	// .. decrypt it ..
	// TODO: requires fix in Conceal & ecc/mock
	// - should be base64 encoded
	// - encrypted response should be a proper (serialized) response object, not only a string, and hence conceal should
	//   return the deserialized response, not a byte array ..
	clearResponse, err := ctx.Reveal([]byte(encryptedResponse))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not decrypt response: %v\n", err)
		os.Exit(1)
	}
	// TODO: create a (single-line) json encoding once we get above a proper response object ...

	payload, err := utils.UnwrapResponse(clearResponse)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	logger.Debugf("Transformed response '%s' to '%s' and write to pipe '%s'", encryptedResponse, string(payload), resultPipeName)
	resultPipeFile.WriteString(fmt.Sprintf("%s\n", payload))
}
