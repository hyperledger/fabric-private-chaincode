/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// This ercc-vscc code is deprecated and will be integrated in ercc with the refactoring

package ercc

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/config"
	"github.com/hyperledger/fabric/core/handlers/decoration"
)

// NewDecorator creates a new decorator
func NewDecorator() decoration.Decorator {
	// TODO bring back init core config
	//common.InitConfig("core")

	// fabric/core/config.GetPath()
	apiKeyFile := config.GetPath("sgx.ias.apiKey.file")
	spidFile := config.GetPath("sgx.ias.spid.file")

	fmt.Printf("api-key: %s\n spid: %s\n", apiKeyFile, spidFile)

	apiKey, err := readApiKeyFromFile(apiKeyFile)
	if err != nil {
		panic("not read api-key from file: " + err.Error())
	}

	spid, err := readSPIDFromFile(spidFile)
	if err != nil {
		panic("Can not read SPID from file: " + err.Error())
	}

	return &decorator{
		apiKey: apiKey,
		spid:   spid,
	}
}

type decorator struct {
	apiKey []byte
	spid   []byte
}

// Decorate decorates a chaincode input by changing it
func (d *decorator) Decorate(proposal *peer.Proposal, input *peer.ChaincodeInput) *peer.ChaincodeInput {
	input.Decorations["SPID"] = d.spid
	input.Decorations["apiKey"] = d.apiKey
	return input
}

func readApiKeyFromFile(apiKeyFile string) ([]byte, error) {
	bytes, err := readFile(apiKeyFile)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func readSPIDFromFile(spidFile string) ([]byte, error) {
	bytes, err := readFile(spidFile)
	if err != nil {
		return nil, err
	}

	return hexSPID2bytes(string(bytes))
}

func hexSPID2bytes(input string) ([]byte, error) {
	spidString := strings.TrimSpace(input)
	spid, err := hex.DecodeString(spidString)
	if err != nil {
		return nil, err
	}

	if len(spid) != 16 {
		return nil, fmt.Errorf("Cannot parse SPID: wrong size")
	}

	return spid, nil
}

func readFile(file string) ([]byte, error) {
	fileCont, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %s, err %s", file, err)
	}
	return fileCont, nil
}
