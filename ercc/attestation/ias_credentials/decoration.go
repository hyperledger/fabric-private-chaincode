/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package main

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric/core/config"
	"github.com/hyperledger/fabric/core/handlers/decoration"
	"github.com/hyperledger/fabric/peer/common"
	"github.com/hyperledger/fabric/protos/peer"
)

// NewDecorator creates a new decorator
func NewDecorator() decoration.Decorator {
	common.InitConfig("core")

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
	return bytes, nil
}

func readFile(file string) ([]byte, error) {
	fileCont, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %s, err %s", file, err)
	}
	return fileCont, nil
}

func main() {
}
