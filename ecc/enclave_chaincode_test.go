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
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/hyperledger-labs/fabric-secure-chaincode/ecc/crypto"
	enc "github.com/hyperledger-labs/fabric-secure-chaincode/ecc/enclave"
	"github.com/hyperledger-labs/fabric-secure-chaincode/ecc/ercc"
	"github.com/hyperledger-labs/fabric-secure-chaincode/ecc/tlcc"
	"github.com/hyperledger-labs/fabric-secure-chaincode/eval/benchmark/executor"
	th "github.com/hyperledger-labs/fabric-secure-chaincode/utils"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func createArgs(stringArgs []string, pk string) [][]byte {
	args_json, _ := json.Marshal(stringArgs)
	return [][]byte{args_json, []byte(pk)}
}

// my tests
func TestEnclaveChaincode_Init(t *testing.T) {
	ecc := createECC()
	stub := shim.NewMockStub("ecc", ecc)

	th.CheckInit(t, stub, nil)
	th.CheckState(t, stub, th.MrEnclaveStateKey, enc.MrEnclave)
}

func TestEnclaveChaincode_Setup(t *testing.T) {
	ecc := createECC()
	stub := shim.NewMockStub("ecc", ecc)

	th.CheckInit(t, stub, nil)
	th.CheckInvoke(t, stub, [][]byte{[]byte("setup"), []byte("ercc"), []byte("mychannel"), []byte("tlcc")})
}

func TestEnclaveChaincode_Invoke(t *testing.T) {
	ecc := createECC()
	stub := shim.NewMockStub("ecc", ecc)

	th.CheckInit(t, stub, nil)
	th.CheckInvoke(t, stub, [][]byte{[]byte("setup"), []byte("ercc"), []byte("mychannel"), []byte("tlcc")})

	stub.State["TestKey"] = []byte("Moin moin")

	clientPk := ""
	args := []string{"create", "MyAuction"}

	th.CheckInvoke(t, stub, createArgs(args, clientPk))
}

type Task struct {
	name     string
	taskID   int
	stub     *shim.MockStub
	args     [][]byte
	callback func(err error)
}

func (t *Task) Invoke() {
	if err := t.doInvoke(); err != nil {
		t.callback(err)
	} else {
		t.callback(nil)
	}
}

func (t *Task) doInvoke() error {
	res := t.stub.MockInvoke("tx"+strconv.Itoa(t.taskID), t.args)
	if res.Status != shim.OK {
		return fmt.Errorf("Invoke %s failed. Reason: %s", t.args, string(res.Message))
	}
	return nil
}

func TestEnclaveChaincode_Invoke_Auction(t *testing.T) {
	ecc := createECC()
	stub := shim.NewMockStub("ecc", ecc)

	pk := ""

	th.CheckInit(t, stub, nil)
	th.CheckInvoke(t, stub, [][]byte{[]byte("setup"), []byte("ercc"), []byte("mychannel"), []byte("tlcc")})

	create_args := createArgs([]string{"create", "MyAuction123"}, pk)

	bid1_args := createArgs([]string{"submit", "MyAuction123", "Bob", "1"}, pk)
	bid2_args := createArgs([]string{"submit", "MyAuction123", "Charly", "7"}, pk)
	bid3_args := createArgs([]string{"submit", "MyAuction123", "France", "5"}, pk)

	close_args := createArgs([]string{"close", "MyAuction123"}, pk)
	eval_args := createArgs([]string{"eval", "MyAuction123"}, pk)

	th.CheckInvoke(t, stub, create_args)

	executor := executor.NewConcurrent("Client", 16)
	executor.Start()
	defer executor.Stop(true)

	var wg sync.WaitGroup
	var mutex sync.RWMutex

	var errs []error
	success := 0

	var bids [][][]byte
	bids = append(bids, bid1_args)
	bids = append(bids, bid2_args)
	bids = append(bids, bid3_args)

	// create tasks
	var tasks []*Task
	for i := 0; i < 10000; i++ {
		myTask := &Task{name: "Rudi",
			taskID: i,
			stub:   stub,
			args:   bids[i%len(bids)],
			callback: func(err error) {
				defer wg.Done()
				mutex.Lock()
				defer mutex.Unlock()
				if err != nil {
					errs = append(errs, err)
				} else {
					success++
				}
			}}
		tasks = append(tasks, myTask)
	}

	numInvocations := len(tasks)
	wg.Add(numInvocations)

	// execute tasks
	for _, task := range tasks {
		// fmt.Printf("Invoke %d\n", task.taskID)
		if err := executor.Submit(task); err != nil {
			panic(fmt.Sprintf("error submitting task: %s", err))
		}
	}

	// Wait for all tasks to complete
	wg.Wait()

	th.CheckInvoke(t, stub, close_args)

	// invoke eval and this time validate response signature
	res := stub.MockInvoke("e1", eval_args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", eval_args, "failed", string(res.Message))
		t.FailNow()
	}

	r := &th.Response{}
	err := json.Unmarshal(res.GetPayload(), r)
	if err != nil {
		t.FailNow()
	}

	string_args := []byte("[\"eval\",\"MyAuction123\"]")

	var writeset [][]byte
	readset := [][]byte{
		[]byte(".somePrefix.MyAuction123.Bob."),
		[]byte(".somePrefix.MyAuction123.Charly."),
		[]byte(".somePrefix.MyAuction123.France."),
		[]byte("MyAuction123"),
	}

	// verify signature
	isValid, err := ecc.verifier.Verify(string_args, r.ResponseData, readset, writeset, r.Signature, r.PublicKey)
	if !isValid {
		t.FailNow()
	}
}

func TestEnclaveChaincode_EncryptedInvoke(t *testing.T) {
	ecc := createECC()
	stub := shim.NewMockStub("ecc", ecc)

	th.CheckInit(t, stub, nil)
	th.CheckInvoke(t, stub, [][]byte{[]byte("setup"), []byte("ercc"), []byte("mychannel"), []byte("tlcc")})

	res := stub.MockInvoke("1", [][]byte{[]byte("getEnclavePk")})
	if res.Status != shim.OK {
		fmt.Println("Invoke getPK failed", string(res.Message))
		t.FailNow()
	}
	r := &th.Response{}
	err := json.Unmarshal(res.GetPayload(), r)
	if err != nil {
		t.FailNow()
	}

	// transform enclave pk
	pk, err := x509.ParsePKIXPublicKey(r.PublicKey)
	if err != nil {
		fmt.Errorf("Failed parsing ecdsa public key [%s]", err)
		t.FailNow()
	}
	enclavePub, ok := pk.(*ecdsa.PublicKey)
	if !ok {
		fmt.Errorf("Verification key is not of type ECDSA")
		t.FailNow()
	}

	// gen my keypair
	priv, pub, err := crypto.GenKeyPair()
	if err != nil {
		fmt.Errorf("Failed to generate key pair [%s]", err)
		t.FailNow()
	}

	// gen shared secret
	key, err := crypto.GenSharedKey(enclavePub, priv)

	plain_args := []string{"create", "MyAuction123"}
	plaintext, _ := json.Marshal(plain_args)
	ciphertext, _ := crypto.Encrypt(plaintext, key[:16])
	fmt.Printf("cipher: \n%s", hex.Dump(ciphertext))

	// transform to sgx pub key format
	pubBytes := make([]byte, 0)
	pubBytes = append(pubBytes, pub.X.Bytes()...)
	pubBytes = append(pubBytes, pub.Y.Bytes()...)

	fmt.Printf("my pk: \n%s", hex.Dump(pubBytes))

	test_args := [][]byte{
		[]byte(base64.StdEncoding.EncodeToString(ciphertext)),
		[]byte(base64.StdEncoding.EncodeToString(pubBytes))}

	for i := 0; i < 1000; i++ {
		res = stub.MockInvoke("1", test_args)
		if res.Status != shim.OK {
			fmt.Println("Invoke getPK failed", string(res.Message))
			t.FailNow()
		}
	}
}

func createECC() *EnclaveChaincode {
	return &EnclaveChaincode{
		erccStub: &ercc.MockEnclaveRegistryStub{},
		tlccStub: &tlcc.MockTLCCStub{},
		enclave:  enc.NewEnclave(),
		verifier: &crypto.ECDSAVerifier{},
	}
}
