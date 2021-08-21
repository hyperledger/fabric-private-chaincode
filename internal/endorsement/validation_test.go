package endorsement

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
	"github.com/hyperledger/fabric-private-chaincode/internal/endorsement/fakes"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/assert"
)

//go:generate counterfeiter -o fakes/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
//lint:ignore U1000 This is just used to generate fake
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o fakes/crypto.go -fake-name CryptoProvider . cryptoProvider
//lint:ignore U1000 This is just used to generate fake
type cryptoProvider interface {
	crypto.CSP
}

func TestReplayReadWrites(t *testing.T) {
	v := &ValidatorImpl{}
	stub := &fakes.ChaincodeStub{}

	// do nothing if no fpcrwset supplied
	err := v.ReplayReadWrites(stub, nil)
	assert.NoError(t, err)

	// error when fabric rw set included
	fpcrwset := &protos.FPCKVSet{}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.Error(t, err)

	// when no reads/write, no GetState/PutState operations are performed
	empty := &kvrwset.KVRWSet{}
	fpcrwset = &protos.FPCKVSet{
		RwSet:           empty,
		ReadValueHashes: nil,
	}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.NoError(t, err)
	assert.Zero(t, stub.PutStateCallCount())
	assert.Zero(t, stub.GetStateCallCount())

	// error when have reads but no hashes
	readA := &kvrwset.KVRead{
		Key: "someKeyA",
	}
	someRWSet := &kvrwset.KVRWSet{
		Reads: []*kvrwset.KVRead{readA},
	}
	fpcrwset = &protos.FPCKVSet{
		RwSet:           someRWSet,
		ReadValueHashes: nil,
	}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.Error(t, err)

	// error when number of reads and hashes not matching
	someHashes := [][]byte{[]byte("some hash"), []byte("another hash")}
	fpcrwset = &protos.FPCKVSet{
		RwSet:           someRWSet,
		ReadValueHashes: someHashes,
	}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.Error(t, err)

	// error when getState for read returns error
	stub.GetStateReturns(nil, fmt.Errorf("some error"))
	someHashes = [][]byte{[]byte("some hash")}
	fpcrwset = &protos.FPCKVSet{
		RwSet:           someRWSet,
		ReadValueHashes: someHashes,
	}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.Error(t, err)

	// error when hash mismatch
	stub.GetStateReturns([]byte("some value"), nil)
	someHashes = [][]byte{[]byte("some hash")}
	fpcrwset = &protos.FPCKVSet{
		RwSet:           someRWSet,
		ReadValueHashes: someHashes,
	}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.Error(t, err)

	// no errors (reads)
	value := []byte("some value")
	stub = &fakes.ChaincodeStub{}
	stub.GetStateReturns(value, nil)
	someHashes = [][]byte{hash(value)}
	fpcrwset = &protos.FPCKVSet{
		RwSet:           someRWSet,
		ReadValueHashes: someHashes,
	}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.NoError(t, err)
	assert.Equal(t, len(someRWSet.Reads), stub.GetStateCallCount())
	for i, r := range someRWSet.Reads {
		k := stub.GetStateArgsForCall(i)
		assert.EqualValues(t, utils.TransformToFPCKey(r.Key), k)
	}

	// error when checking writeset and putstate returns error
	stub = &fakes.ChaincodeStub{}
	stub.PutStateReturns(fmt.Errorf("some error"))
	writeA := &kvrwset.KVWrite{
		Key:   "someKey",
		Value: []byte("some value"),
	}
	someRWSet = &kvrwset.KVRWSet{
		Writes: []*kvrwset.KVWrite{writeA},
	}
	fpcrwset = &protos.FPCKVSet{
		RwSet: someRWSet,
	}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.Error(t, err)

	// no error (writes)
	stub = &fakes.ChaincodeStub{}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.NoError(t, err)
	for i, w := range someRWSet.Writes {
		k, val := stub.PutStateArgsForCall(i)
		assert.EqualValues(t, utils.TransformToFPCKey(w.Key), k)
		assert.EqualValues(t, w.Value, val)
	}

	// no error (writes) with comp keys
	expectedFabricCompKey := "\x00some\x00Key\x00"
	writeCompKey := &kvrwset.KVWrite{
		Key:   ".some.Key.",
		Value: []byte("some value"),
	}
	someRWSet = &kvrwset.KVRWSet{
		Writes: []*kvrwset.KVWrite{writeCompKey},
	}
	fpcrwset = &protos.FPCKVSet{
		RwSet: someRWSet,
	}
	stub = &fakes.ChaincodeStub{}
	stub.CreateCompositeKeyReturns(expectedFabricCompKey, nil)
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.NoError(t, err)
	k, val := stub.PutStateArgsForCall(0)
	assert.EqualValues(t, expectedFabricCompKey, k)
	assert.EqualValues(t, writeCompKey.Value, val)

	// error when rangequery
	someRWSet = &kvrwset.KVRWSet{
		RangeQueriesInfo: []*kvrwset.RangeQueryInfo{{
			StartKey: "start",
			EndKey:   "end",
		}},
	}
	fpcrwset = &protos.FPCKVSet{
		RwSet: someRWSet,
	}
	err = v.ReplayReadWrites(stub, fpcrwset)
	assert.Error(t, err)
}

func TestValidate(t *testing.T) {
	// TODO
	c := &fakes.CryptoProvider{}
	v := &ValidatorImpl{csp: c}

	scr := &protos.SignedChaincodeResponseMessage{}

	// error when no signature in SignedChaincodeResponse Message
	err := v.Validate(scr, nil)
	assert.Error(t, err)

	// error when no chaincode response included
	scr = &protos.SignedChaincodeResponseMessage{
		Signature: []byte("some signature"),
	}
	err = v.Validate(scr, nil)
	assert.Error(t, err)

	// error when no enclave verification key included in attestedData
	scr = &protos.SignedChaincodeResponseMessage{
		Signature:                []byte("some signature"),
		ChaincodeResponseMessage: []byte("some message"),
	}
	at := &protos.AttestedData{}
	err = v.Validate(scr, at)
	assert.Error(t, err)

	// error when signature verification failed
	at = &protos.AttestedData{
		EnclaveVk: []byte("some key"),
	}
	c.VerifyMessageReturns(fmt.Errorf("some error"))
	err = v.Validate(scr, at)
	assert.Error(t, err)

	// error when input hash mismatch detected
	expectedChaincodeRequestMsg := []byte("someMsg")
	expectedHash := sha256.Sum256([]byte("hashMismatch!!!"))
	response := createChaincodeResponseMessage(expectedChaincodeRequestMsg, expectedHash[:])
	scr = &protos.SignedChaincodeResponseMessage{
		Signature:                []byte("some signature"),
		ChaincodeResponseMessage: utils.MarshalOrPanic(response),
	}
	c.VerifyMessageReturns(nil)
	err = v.Validate(scr, at)
	assert.Error(t, err)

	// no errors
	expectedHash = sha256.Sum256(expectedChaincodeRequestMsg)
	response = createChaincodeResponseMessage(expectedChaincodeRequestMsg, expectedHash[:])
	scr = &protos.SignedChaincodeResponseMessage{
		Signature:                []byte("some signature"),
		ChaincodeResponseMessage: utils.MarshalOrPanic(response),
	}
	c.VerifyMessageReturns(nil)
	err = v.Validate(scr, at)
	assert.NoError(t, err)
}

func createChaincodeResponseMessage(chaincodeRequest []byte, chaincodeRequestHash []byte) *protos.ChaincodeResponseMessage {
	chdr := &common.ChannelHeader{
		Type:      int32(common.HeaderType_ENDORSER_TRANSACTION),
		ChannelId: "someChannelId",
	}

	header := &common.Header{
		ChannelHeader:   protoutil.MarshalOrPanic(chdr),
		SignatureHeader: nil,
	}

	input := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type:        0,
			ChaincodeId: nil,
			Input: &peer.ChaincodeInput{
				Args: [][]byte{[]byte("__invoke"), []byte(base64.StdEncoding.EncodeToString(chaincodeRequest))},
			},
		},
	}

	payload := &peer.ChaincodeProposalPayload{
		Input:        protoutil.MarshalOrPanic(input),
		TransientMap: nil,
	}

	proposal := &peer.Proposal{
		Header:    protoutil.MarshalOrPanic(header),
		Payload:   protoutil.MarshalOrPanic(payload),
		Extension: nil,
	}

	return &protos.ChaincodeResponseMessage{
		EncryptedResponse: []byte("someEncryptedResponse"),
		FpcRwSet:          nil,
		Proposal: &peer.SignedProposal{
			ProposalBytes: protoutil.MarshalOrPanic(proposal),
			Signature:     nil,
		},
		ChaincodeRequestMessageHash: chaincodeRequestHash,
		EnclaveId:                   "someEnclaveId",
	}
}

func hash(v []byte) []byte {
	h := sha256.New()
	h.Write(v)
	return h.Sum(nil)
}
