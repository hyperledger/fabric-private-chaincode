/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package registry_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-private-chaincode/ercc/registry"
	"github.com/hyperledger/fabric-private-chaincode/ercc/registry/fakes"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

//go:generate counterfeiter -o fakes/transaction.go -fake-name TransactionContext . transactionContext
//lint:ignore U1000 This is just used to generate fake
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate counterfeiter -o fakes/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
//lint:ignore U1000 This is just used to generate fake
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o fakes/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
//lint:ignore U1000 This is just used to generate fake
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

//go:generate counterfeiter -o fakes/verifier.go -fake-name CredentialVerifier . credentialVerifier
//lint:ignore U1000 This is just used to generate fake
type credentialVerifier interface {
	attestation.Verifier
}

//go:generate counterfeiter -o fakes/evaluator.go -fake-name IdentityEvaluator . identityEvaluator
//lint:ignore U1000 This is just used to generate fake
type identityEvaluator interface {
	utils.IdentityEvaluatorInterface
}

var (
	mrenclave   = `98aed61c91f258a37c68ed4943297695647ec7bbe6008cc111b0a12650ebeb91`
	channelId   = "MY_TEST_CHANNEL"
	chaincodeId = "SOME_CHAINCODE_PKG_ID"
	enclaveId   = "some enclave id"
	someMspId   = "some org"
)

func toBase64(credentials *protos.Credentials) string {
	credentialBytes := protoutil.MarshalOrPanic(credentials)
	return base64.StdEncoding.EncodeToString(credentialBytes)
}

func TestRegisterEnclave(t *testing.T) {

	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	verifier := &fakes.CredentialVerifier{}
	verifier.VerifyCredentialsReturns(nil)

	id := &fakes.IdentityEvaluator{}

	ercc := registry.Contract{}
	ercc.Verifier = verifier
	ercc.IEvaluator = id

	err := ercc.RegisterEnclave(transactionContext, "")
	require.EqualError(t, err, "invalid credential bytes: credential input empty")

	err = ercc.RegisterEnclave(transactionContext, "some bytes")
	require.Contains(t, err.Error(), "invalid credential bytes")

	err = ercc.RegisterEnclave(transactionContext, toBase64(
		&protos.Credentials{
			Evidence: []byte("some bytes"),
		}))
	require.Contains(t, err.Error(), "attested data is empty")

	err = ercc.RegisterEnclave(transactionContext, toBase64(
		&protos.Credentials{
			SerializedAttestedData: &anypb.Any{},
		}))
	require.Contains(t, err.Error(), "evidence is empty")

	err = ercc.RegisterEnclave(transactionContext, toBase64(
		&protos.Credentials{
			SerializedAttestedData: &anypb.Any{},
			Evidence:               []byte("some bytes"),
		}))
	require.Contains(t, err.Error(), "invalid attested data message")

	chaincodeStub.GetChannelIDReturns("ANOTHER_CHANNEL")
	serializedAttestedData, _ := anypb.New(
		&protos.AttestedData{
			EnclaveVk: []byte("enclaveVKString"),
			CcParams: &protos.CCParameters{
				ChannelId: "WRONG_CHANNEL",
			},
		})
	err = ercc.RegisterEnclave(transactionContext, toBase64(
		&protos.Credentials{
			Evidence:               []byte("some mock evidence"),
			SerializedAttestedData: serializedAttestedData,
		}))
	require.EqualError(t, err, "wrong channel! expected=ANOTHER_CHANNEL, actual=WRONG_CHANNEL")

	chaincodeStub.GetChannelIDReturns(channelId)
	chaincodeStub.InvokeChaincodeReturns(shim.Error("no chaincode definition exists"))
	serializedAttestedData, _ = anypb.New(
		&protos.AttestedData{
			EnclaveVk: []byte("enclaveVKString"),
			CcParams: &protos.CCParameters{
				ChannelId: channelId,
			},
		})
	err = ercc.RegisterEnclave(transactionContext, toBase64(
		&protos.Credentials{
			Evidence:               []byte("some mock evidence"),
			SerializedAttestedData: serializedAttestedData,
		}))
	require.Contains(t, err.Error(), "cannot get chaincode definition")

	// create mock lifecycle chaincode
	chaincodeStub.InvokeChaincodeReturns(shim.Success(protoutil.MarshalOrPanic(
		&lifecycle.QueryChaincodeDefinitionResult{
			Version: mrenclave,
		})))

	serializedAttestedData, _ = anypb.New(
		&protos.AttestedData{
			EnclaveVk: []byte("enclaveVKString"),
			CcParams: &protos.CCParameters{
				ChaincodeId: chaincodeId,
				Version:     "WRONG_MRENCLAVE",
				ChannelId:   channelId,
			},
		})
	credentialBase64 := toBase64(&protos.Credentials{
		Evidence:               []byte("some mock evidence"),
		SerializedAttestedData: serializedAttestedData,
	})
	err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	require.EqualError(t, err, "mrenclave does not match chaincode definition")

	chaincodeStub.InvokeChaincodeReturns(shim.Success(protoutil.MarshalOrPanic(
		&lifecycle.QueryChaincodeDefinitionResult{
			Version:  mrenclave,
			Sequence: 1,
		})))

	serializedAttestedData, _ = anypb.New(
		&protos.AttestedData{
			EnclaveVk: []byte("enclaveVKString"),
			CcParams: &protos.CCParameters{
				ChaincodeId: chaincodeId,
				Version:     mrenclave,
				ChannelId:   channelId,
				Sequence:    666,
			},
		})
	credentialBase64 = toBase64(&protos.Credentials{
		Evidence:               []byte("some mock evidence"),
		SerializedAttestedData: serializedAttestedData,
	})
	err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	require.EqualError(t, err, "sequence does not match chaincode definition")

	chaincodeStub.InvokeChaincodeReturns(shim.Success(protoutil.MarshalOrPanic(
		&lifecycle.QueryChaincodeDefinitionResult{
			Version:  mrenclave,
			Sequence: 1,
		})))
	verifier.VerifyCredentialsReturns(fmt.Errorf("evidence invalid"))

	serializedAttestedData, _ = anypb.New(
		&protos.AttestedData{
			EnclaveVk: []byte("enclaveVKString"),
			CcParams: &protos.CCParameters{
				ChaincodeId: chaincodeId,
				Version:     mrenclave,
				ChannelId:   channelId,
				Sequence:    1,
			},
		})
	credentialBase64 = toBase64(&protos.Credentials{
		Evidence:               []byte("some mock evidence"),
		SerializedAttestedData: serializedAttestedData,
	})
	err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	require.EqualError(t, err, "evidence verification failed: evidence invalid")

	verifier.VerifyCredentialsReturns(nil)
	err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	require.EqualError(t, err, "host params are empty")

	serializedAttestedData, _ = anypb.New(
		&protos.AttestedData{
			EnclaveVk: []byte("enclaveVKString"),
			CcParams: &protos.CCParameters{
				ChaincodeId: chaincodeId,
				Version:     mrenclave,
				ChannelId:   channelId,
				Sequence:    1,
			},
			HostParams: &protos.HostParameters{
				PeerMspId: someMspId,
			},
		})
	credentialBase64 = toBase64(&protos.Credentials{
		Evidence:               []byte("some mock evidence"),
		SerializedAttestedData: serializedAttestedData,
	})
	//id.EvaluateIdentityReturns(fmt.Errorf("peer not a valid endorser"))
	//err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	//require.EqualError(t, err, "identity does not satisfy endorsement policy: peer not a valid endorser")

	id.EvaluateCreatorIdentityReturns(nil)
	chaincodeStub.GetCreatorReturns(nil, fmt.Errorf("cannot get creator"))
	err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	require.EqualError(t, err, "cannot get creator")

	id.EvaluateCreatorIdentityReturns(fmt.Errorf("msp does not match"))
	chaincodeStub.GetCreatorReturns([]byte("fake creator"), nil)
	err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	require.EqualError(t, err, "creator identity evaluation failed: msp does not match")

	id.EvaluateCreatorIdentityReturns(nil)
	chaincodeStub.CreateCompositeKeyReturns("someString", fmt.Errorf("cannot create composite key"))
	err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	require.EqualError(t, err, "cannot create composite key")

	chaincodeStub.CreateCompositeKeyReturns("someKey", nil)
	chaincodeStub.PutStateReturns(fmt.Errorf("some put state error"))
	err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	require.EqualError(t, err, "cannot store credentials: some put state error")

	chaincodeStub.PutStateReturns(nil)
	err = ercc.RegisterEnclave(transactionContext, credentialBase64)
	require.NoError(t, err)
}

func TestQueryListEnclaveCredentials(t *testing.T) {
	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	ercc := registry.Contract{}

	chaincodeStub.GetStateByPartialCompositeKeyReturns(nil, fmt.Errorf("some error"))
	resp, err := ercc.QueryListEnclaveCredentials(transactionContext, chaincodeId)
	require.Empty(t, resp)
	require.EqualError(t, err, "some error")

	stateQueryIterator := &fakes.StateQueryIterator{}
	stateQueryIterator.HasNextReturns(false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(stateQueryIterator, nil)
	resp, err = ercc.QueryListEnclaveCredentials(transactionContext, chaincodeId)
	require.Empty(t, resp)
	require.NoError(t, err)

	stateQueryIterator = &fakes.StateQueryIterator{}
	stateQueryIterator.HasNextReturnsOnCall(0, true)
	stateQueryIterator.NextReturns(nil, fmt.Errorf("some query error"))
	chaincodeStub.GetStateByPartialCompositeKeyReturns(stateQueryIterator, nil)
	resp, err = ercc.QueryListEnclaveCredentials(transactionContext, chaincodeId)
	require.Empty(t, resp)
	require.EqualError(t, err, "some query error")

	stateQueryIterator = &fakes.StateQueryIterator{}
	stateQueryIterator.HasNextReturnsOnCall(0, true)
	stateQueryIterator.HasNextReturnsOnCall(1, false)
	value := []byte(base64.StdEncoding.EncodeToString([]byte("some item")))
	stateQueryIterator.NextReturns(&queryresult.KV{Value: value}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(stateQueryIterator, nil)
	resp, err = ercc.QueryListEnclaveCredentials(transactionContext, chaincodeId)
	require.Contains(t, resp, base64.StdEncoding.EncodeToString([]byte("some item")))
	require.NoError(t, err)

	stateQueryIterator = &fakes.StateQueryIterator{}
	stateQueryIterator.HasNextReturnsOnCall(0, true)
	stateQueryIterator.HasNextReturnsOnCall(1, true)
	stateQueryIterator.HasNextReturnsOnCall(2, false)
	value1 := []byte(base64.StdEncoding.EncodeToString([]byte("some item-1")))
	value2 := []byte(base64.StdEncoding.EncodeToString([]byte("some item-2")))
	stateQueryIterator.NextReturnsOnCall(0, &queryresult.KV{Value: value1}, nil)
	stateQueryIterator.NextReturnsOnCall(1, &queryresult.KV{Value: value2}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(stateQueryIterator, nil)
	resp, err = ercc.QueryListEnclaveCredentials(transactionContext, chaincodeId)
	require.Contains(t, resp, base64.StdEncoding.EncodeToString([]byte("some item-1")))
	require.Contains(t, resp, base64.StdEncoding.EncodeToString([]byte("some item-2")))
	require.Equal(t, len(resp), 2)
	require.NoError(t, err)
}

func TestQueryEnclaveCredentials(t *testing.T) {
	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	ercc := registry.Contract{}

	chaincodeStub.CreateCompositeKeyReturns("some key", fmt.Errorf("some error"))
	resp, err := ercc.QueryEnclaveCredentials(transactionContext, chaincodeId, enclaveId)
	require.Empty(t, resp)
	require.EqualError(t, err, "some error")
	objType, attr := chaincodeStub.CreateCompositeKeyArgsForCall(0)
	require.Equal(t, objType, "namespaces/credentials")
	require.Contains(t, attr, chaincodeId)
	require.Contains(t, attr, enclaveId)

	chaincodeStub.CreateCompositeKeyReturns("some key", nil)
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("get state error"))
	resp, err = ercc.QueryEnclaveCredentials(transactionContext, chaincodeId, enclaveId)
	require.Empty(t, resp)
	require.EqualError(t, err, "get state error")
	k := chaincodeStub.GetStateArgsForCall(0)
	require.Equal(t, k, "some key")

	chaincodeStub.GetStateReturns([]byte("credentialBytes"), nil)
	resp, err = ercc.QueryEnclaveCredentials(transactionContext, chaincodeId, enclaveId)
	require.Equal(t, resp, "credentialBytes")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	resp, err = ercc.QueryEnclaveCredentials(transactionContext, chaincodeId, enclaveId)
	require.Empty(t, resp)
	require.NoError(t, err)
}
