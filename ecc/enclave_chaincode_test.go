/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ecc

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

	"github.com/hyperledger-labs/fabric-private-chaincode/ecc/crypto"
	th "github.com/hyperledger-labs/fabric-private-chaincode/internal/testing"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/testing/executor"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
)

func createArgs(stringArgs []string, pk string) [][]byte {
	args_json, _ := json.Marshal(stringArgs)
	return [][]byte{args_json, []byte(pk)}
}

// my tests
func TestEnclaveChaincode_Init(t *testing.T) {
	ecc := CreateMockedECC()
	stub := shimtest.NewMockStub("ecc", ecc)

	th.CheckInit(t, stub, nil)
	// th.CheckState(t, stub, th.MrEnclaveStateKey, enc.MrEnclave)
}

func TestEnclaveChaincode_Setup(t *testing.T) {
	ecc := CreateMockedECC()
	stub := shimtest.NewMockStub("ecc", ecc)

	th.CheckInit(t, stub, nil)
	th.CheckInvoke(t, stub, [][]byte{[]byte("__setup"), []byte("ercc"), []byte("mychannel"), []byte("tlcc")})
}

func TestEnclaveChaincode_Invoke(t *testing.T) {
	ecc := CreateMockedECC()
	stub := shimtest.NewMockStub("ecc", ecc)

	th.CheckInit(t, stub, nil)
	th.CheckInvoke(t, stub, [][]byte{[]byte("__setup"), []byte("ercc"), []byte("mychannel"), []byte("tlcc")})

	stub.State["TestKey"] = []byte("Moin moin")

	clientPk := ""
	args := []string{"create", "MyAuction"}

	th.CheckInvoke(t, stub, createArgs(args, clientPk))
}

type Task struct {
	name     string
	taskID   int
	stub     *shimtest.MockStub
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
	ecc := CreateMockedECC()
	stub := shimtest.NewMockStub("ecc", ecc)

	pk := ""

	th.CheckInit(t, stub, nil)
	th.CheckInvoke(t, stub, [][]byte{[]byte("__setup"), []byte("ercc"), []byte("mychannel"), []byte("tlcc")})

	create_args := createArgs([]string{"create", "MyAuction123"}, pk)

	bid1_args := createArgs([]string{"submit", "MyAuction123", "Bob", "1"}, pk)
	bid2_args := createArgs([]string{"submit", "MyAuction123", "Charly", "7"}, pk)
	bid3_args := createArgs([]string{"submit", "MyAuction123", "France", "5"}, pk)

	close_args := createArgs([]string{"close", "MyAuction123"}, pk)
	eval_args := createArgs([]string{"eval", "MyAuction123"}, pk)

	th.CheckInvoke(t, stub, create_args)

	// note that shim.MockStub is not thread-safe! Therefore we better use a single executor here
	executor := executor.NewConcurrent("Client", 1)
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

	r := &utils.Response{}
	err := json.Unmarshal(res.GetPayload(), r)
	if err != nil {
		t.FailNow()
	}

	txType := []byte("invoke")
	string_args := []byte("[\"eval\",\"MyAuction123\"]")

	var writeset [][]byte
	readset := [][]byte{
		[]byte(".somePrefix.MyAuction123.Bob."),
		[]byte(".somePrefix.MyAuction123.Charly."),
		[]byte(".somePrefix.MyAuction123.France."),
		[]byte("MyAuction123"),
	}

	// verify signature
	isValid, err := ecc.verifier.Verify(txType, string_args, r.ResponseData, readset, writeset, r.Signature, r.PublicKey)
	if !isValid {
		t.FailNow()
	}
}

func TestEnclaveChaincode_EncryptedInvoke(t *testing.T) {
	ecc := CreateMockedECC()
	stub := shimtest.NewMockStub("ecc", ecc)

	th.CheckInit(t, stub, nil)
	th.CheckInvoke(t, stub, [][]byte{[]byte("__setup"), []byte("ercc"), []byte("mychannel"), []byte("tlcc")})

	res := stub.MockInvoke("1", [][]byte{[]byte("__getEnclavePk")})
	if res.Status != shim.OK {
		fmt.Println("Invoke getPK failed", string(res.Message))
		t.FailNow()
	}
	r := &utils.Response{}
	err := json.Unmarshal(res.GetPayload(), r)
	if err != nil {
		t.FailNow()
	}

	// transform enclave pk
	pk, err := x509.ParsePKIXPublicKey(r.PublicKey)
	if err != nil {
		t.Errorf("failed parsing ecdsa public key [%s]", err)
	}
	enclavePub, ok := pk.(*ecdsa.PublicKey)
	if !ok {
		t.Error("verification key is not of type ECDSA")
	}

	// gen my keypair
	priv, pub, err := crypto.GenKeyPair()
	if err != nil {
		t.Errorf("failed to generate key pair [%s]", err)
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
