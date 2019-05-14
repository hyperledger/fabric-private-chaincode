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
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation/mock"
	th "github.com/hyperledger-labs/fabric-private-chaincode/utils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const enclavePK = `MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEE9lPD9QkW9oxWlFvwABrmseYAVvoBvvmTt3jzV0sdASR2KDDQPvz8EcyqfomEOTwSz7E+mISktMxYqofRr+4Yw==`
const enclavePkHash = `qpEqqBaEkNz9bTO77QK8+CLbvaEN1NATs7ajRTzq70k=`
const quote = `AgAAAG4NAAAEAAQAAAAAACVC+Q1jMSwdovbiGHbw44nMDb+CvAvF0FJF/38NWjOqAgIC/wEBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABwAAAAAAAAAHAAAAAAAAAJiu1hyR8lijfGjtSUMpdpVkfse75gCMwRGwoSZQ6+uRAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACD1xnnferKFHD2uvYqTXdDA8iZ22kCD5xw7h38CMfOngAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD3TvjLWa36sT/kCIRYXhtYoRQ61x2u48Q16bzoq8w6egAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAqAIAAFXdt6ObnofTxKVhK9Eafot/LsUGgr4546W34JUey7aqo9b6mpeP3X6W/DMIc1JbIXpFd5+mHWP+R7swgSEYNg+RUVbjkZ38nOJHzIl0E7Dgxs8X8iilH+hxpcPiYQphpIcBS5NUCmDn6Wsz/I+Dbpbt3e2G74WFPLHqDn+JHva5vtaYHd7cAfmPIhZZXXMCQJ8um5Jcer4L16VOugt8LEE0i3FqLb0khMYUHmEsqWuh1Fss5bNUuRDqotz6XTBq0uQ+nCfzv9ZsT2CDihsuQTzgU0BiZZuf06Aw9NQdywg+vTZoqyWw0Ca/jsAt+OpbQeQzDoH3HAvnaRvRByozHqKQ1Z83vVny2DQPVwWm6hxIEUCDVE2A/fkbo+UjR12fD8XWUw3xXfd6Dob9N2gBAAB1NRKH8uhAp94KvF/EF76xtBOYnlpAkbv4pYsmJfWkt0CtKtt/lvMQqkmZwSi8LQ93XBiAdVEKt255ycfFxcmAHPFPrjwHMb0/5wKNXa9vyBlgJ63tU/8U1JxujZ6QdS05xiQbKb+l2y6Nm++iw1Ba7BBJgQR+xDBud/VMjjLI3/nMlA9JTpVw9sSTsWdqHzA4bJm2P7fxkxL4wUYe6w+1uWGnT8XFwuJOfw1bUKZWlGZCOe8iLiPmDOmKUegpiLy0wY73gk+5bJhq1L8b4EXJMoSVoS4JgzYajh8oEBaUheiR4ze8sD9KuF0y+dfQklcMKdONyXMcI8QcZfj19iQy2FvXY8Ca0AoBkQMk4bn49e19ePChDUhrk7ynGGy5d9Wo8g3aNZLNWol5LuwCduTYv83xbHeKDkEsvk23m5NiXlVnDo6Pwu+32w57sX4K4CcojQZvJRYfFUuRCoN05TY0oJ0qvvZ1pAEAQAuBfOucbX6QZZ4qPcMR`

func TestEnclaveRegistry_Init(t *testing.T) {
	ercc := NewTestErcc()
	stub := shim.NewMockStub("ercc", ercc)

	// Init
	th.CheckInit(t, stub, [][]byte{})
}

func TestEnclaveRegistry_Register(t *testing.T) {
	ercc := NewTestErcc()
	stub := shim.NewMockStub("ercc", ercc)
	stub.Decorations["apiKey"] = []byte(mock.MOCK_ApiKey)

	// Init
	th.CheckInit(t, stub, [][]byte{})

	// invoke registerEnclave
	th.CheckInvoke(t, stub, [][]byte{[]byte("registerEnclave"), []byte(enclavePK), []byte(quote)})
	th.CheckStateNotNull(t, stub, enclavePkHash)
}

func TestEnclaveRegistry_GetAttestationReport(t *testing.T) {
	ercc := NewTestErcc()
	stub := shim.NewMockStub("ercc", ercc)
	stub.Decorations["apiKey"] = []byte(mock.MOCK_ApiKey)

	// Init
	th.CheckInit(t, stub, [][]byte{})

	// invoke registerEnclave
	th.CheckInvoke(t, stub, [][]byte{[]byte("registerEnclave"), []byte(enclavePK), []byte(quote)})
	th.CheckStateNotNull(t, stub, enclavePkHash)

	th.CheckQueryNotNull(t, stub, [][]byte{[]byte("getAttestationReport"), []byte(enclavePkHash)})
}

func TestEnclaveRegistry_GetSPID(t *testing.T) {
	ercc := NewTestErcc()
	stub := shim.NewMockStub("ercc", ercc)
	stub.Decorations["SPID"] = mock.MOCK_SPID[:]

	// Init
	th.CheckInit(t, stub, [][]byte{})

	// invoke getSPID
	th.CheckQuery(t, stub, [][]byte{[]byte("getSPID")}, string(mock.MOCK_SPID[:]))
}

func TestEnclaveRegistry_PEM(t *testing.T) {
	asBase := base64.StdEncoding.EncodeToString([]byte(attestation.IntelPubPEM))
	andBack, _ := base64.StdEncoding.DecodeString(asBase)

	block, _ := pem.Decode(andBack)
	if block == nil {
		t.Fatalf("IntelPubPEM is invalid: failed to parse PEM block containing the public key")
	}
	_, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("IntelPubPEM is invalid: ParsePKIXPublicKey")
	}
	t.Log("Success")
}
