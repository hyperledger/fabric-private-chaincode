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
	"encoding/pem"
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
	certFile := config.GetPath("sgx.ias.cert.file")
	keyFile := config.GetPath("sgx.ias.key.file")
	spidFile := config.GetPath("sgx.ias.spid.file")

	fmt.Printf("cert: %s\n key: %s\n spid: %s\n", certFile, keyFile, spidFile)

	certPEM, err := readPemFromFile(certFile)
	if err != nil {
		panic("not read Cert from file: " + err.Error())
	}

	keyPEM, err := readPemFromFile(keyFile)
	if err != nil {
		panic("not read Cert from file: " + err.Error())
	}

	spid, err := readSPIDFromFile(spidFile)
	if err != nil {
		panic("Can not read SPID from file: " + err.Error())
	}

	return &decorator{
		certPEM: certPEM,
		keyPEM:  keyPEM,
		spid:    spid,
	}
}

type decorator struct {
	certPEM []byte
	keyPEM  []byte
	spid    []byte
}

// Decorate decorates a chaincode input by changing it
func (d *decorator) Decorate(proposal *peer.Proposal, input *peer.ChaincodeInput) *peer.ChaincodeInput {
	input.Decorations["SPID"] = d.spid
	input.Decorations["certPEM"] = d.certPEM
	input.Decorations["keyPEM"] = d.keyPEM
	return input
}

func readSPIDFromFile(spidFile string) ([]byte, error) {
	bytes, err := readFile(spidFile)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func readPemFromFile(file string) ([]byte, error) {
	bytes, err := readFile(file)
	if err != nil {
		return nil, err
	}

	b, _ := pem.Decode(bytes)
	if b == nil { // TODO: also check that the type is what we expect (cert vs key..)
		return nil, fmt.Errorf("No pem content for file %s", file)
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
