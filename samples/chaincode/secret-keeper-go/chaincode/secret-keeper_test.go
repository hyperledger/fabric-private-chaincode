package chaincode

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/ercc/registry/fakes"
	"github.com/stretchr/testify/require"
)

func TestWrongSignature(t *testing.T) {
	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	secretKeeper := SecretKeeper{}
	err := secretKeeper.InitSecretKeeper(transactionContext)
	require.NoError(t, err)

	_, authSetByte := chaincodeStub.PutStateArgsForCall(0)

	chaincodeStub.GetStateReturns(authSetByte, nil)

	falseSig := "falseSignature"
	fakeSecret := "fakeSecret"
	err = secretKeeper.AddUser(transactionContext, falseSig, falseSig)
	require.EqualError(t, err, "User are not allowed to perform this action.")

	err = secretKeeper.RemoveUser(transactionContext, falseSig, falseSig)
	require.EqualError(t, err, "User are not allowed to perform this action.")

	err = secretKeeper.LockSecret(transactionContext, falseSig, fakeSecret)
	require.EqualError(t, err, "User are not allowed to perform this action.")

	secret, err := secretKeeper.RevealSecret(transactionContext, falseSig)
	require.EqualError(t, err, "User are not allowed to perform this action.")
	require.Nil(t, secret)
}

func TestAddUserFlow(t *testing.T) {
	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	secretKeeper := SecretKeeper{}
	err := secretKeeper.InitSecretKeeper(transactionContext)
	require.NoError(t, err)

	_, authSetByte := chaincodeStub.PutStateArgsForCall(0)
	_, _ = chaincodeStub.PutStateArgsForCall(1) // get default secret

	chaincodeStub.GetStateReturns(authSetByte, nil)

	aliceSig := "Alice"
	evePubKey := "Eve"

	// check if authlist not contains eve
	var authSet AuthSet
	err = json.Unmarshal(authSetByte, &authSet)
	require.NoError(t, err)
	_, exist := authSet.Pubkey[evePubKey]
	require.False(t, exist)

	err = secretKeeper.AddUser(transactionContext, aliceSig, evePubKey)
	require.NoError(t, err)

	// check if authlist contains eve.
	_, authSetByte2 := chaincodeStub.PutStateArgsForCall(2)
	var authSet2 AuthSet
	err = json.Unmarshal(authSetByte2, &authSet2)
	require.NoError(t, err)
	_, exist = authSet2.Pubkey[evePubKey]
	require.True(t, exist)
}

func TestRemoveUser(t *testing.T) {
	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	secretKeeper := SecretKeeper{}
	err := secretKeeper.InitSecretKeeper(transactionContext)
	require.NoError(t, err)

	_, authSetByte := chaincodeStub.PutStateArgsForCall(0)
	_, _ = chaincodeStub.PutStateArgsForCall(1) // get default secret

	chaincodeStub.GetStateReturns(authSetByte, nil)

	aliceSig := "Alice"
	bobPubKey := "Bob"

	// check if authlist contains bob.
	var authSet AuthSet
	err = json.Unmarshal(authSetByte, &authSet)
	require.NoError(t, err)
	_, exist := authSet.Pubkey[bobPubKey]
	require.True(t, exist)

	err = secretKeeper.RemoveUser(transactionContext, aliceSig, bobPubKey)
	require.NoError(t, err)

	// check if authlist doesn't contain bob anymore.
	_, authSetByte2 := chaincodeStub.PutStateArgsForCall(2)
	var authSet2 AuthSet
	err = json.Unmarshal(authSetByte2, &authSet2)
	require.NoError(t, err)
	_, exist = authSet2.Pubkey[bobPubKey]
	require.False(t, exist)
}

func TestLockSecret(t *testing.T) {
	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	secretKeeper := SecretKeeper{}
	err := secretKeeper.InitSecretKeeper(transactionContext)
	require.NoError(t, err)

	_, authSetByte := chaincodeStub.PutStateArgsForCall(0)
	_, _ = chaincodeStub.PutStateArgsForCall(1) // get default secret

	chaincodeStub.GetStateReturns(authSetByte, nil)

	aliceSig := "Alice"
	newSecret := "newSecret"

	err = secretKeeper.LockSecret(transactionContext, aliceSig, newSecret)
	require.NoError(t, err)

	// check secret key value.
	_, secretByte := chaincodeStub.PutStateArgsForCall(2)
	var secret Secret
	err = json.Unmarshal(secretByte, &secret)
	require.NoError(t, err)
	require.EqualValues(t, secret.Value, newSecret)
}

func TestRevealSecret(t *testing.T) {
	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	secretKeeper := SecretKeeper{}
	err := secretKeeper.InitSecretKeeper(transactionContext)
	require.NoError(t, err)

	_, authSetByte := chaincodeStub.PutStateArgsForCall(0)
	_, defaultSecretByte := chaincodeStub.PutStateArgsForCall(1)

	aliceSig := "Alice"
	var defaultSecret Secret
	err = json.Unmarshal(defaultSecretByte, &defaultSecret)
	require.NoError(t, err)

	// check the return value equal with the secret in test.
	chaincodeStub.GetStateReturnsOnCall(0, authSetByte, nil)
	chaincodeStub.GetStateReturnsOnCall(1, defaultSecretByte, nil)
	secret, err := secretKeeper.RevealSecret(transactionContext, aliceSig)
	require.NoError(t, err)
	require.EqualValues(t, secret.Value, defaultSecret.Value)
}

func TestNormalBehavior(t *testing.T) {
	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	secretKeeper := SecretKeeper{}
	err := secretKeeper.InitSecretKeeper(transactionContext)
	require.NoError(t, err)

	_, authSetByte := chaincodeStub.PutStateArgsForCall(0)
	_, secretByte := chaincodeStub.PutStateArgsForCall(1)

	aliceSig := "Alice"
	bobSig := "Bob"
	eveSig := "Eve"
	newSecret := "NewSecret"
	newSecret2 := "SecretWithoutAlice"

	chaincodeStub.GetStateReturnsOnCall(0, authSetByte, nil)
	chaincodeStub.GetStateReturnsOnCall(1, authSetByte, nil)
	err = secretKeeper.AddUser(transactionContext, aliceSig, eveSig)
	require.NoError(t, err)
	_, authSetByte = chaincodeStub.PutStateArgsForCall(2)

	chaincodeStub.GetStateReturnsOnCall(2, authSetByte, nil)
	err = secretKeeper.LockSecret(transactionContext, eveSig, newSecret)
	require.NoError(t, err)
	_, secretByte = chaincodeStub.PutStateArgsForCall(3)

	chaincodeStub.GetStateReturnsOnCall(3, authSetByte, nil)
	chaincodeStub.GetStateReturnsOnCall(4, secretByte, nil)
	secret, err := secretKeeper.RevealSecret(transactionContext, aliceSig)
	require.NoError(t, err)
	require.EqualValues(t, secret.Value, newSecret)

	chaincodeStub.GetStateReturnsOnCall(5, authSetByte, nil)
	chaincodeStub.GetStateReturnsOnCall(6, authSetByte, nil)
	err = secretKeeper.RemoveUser(transactionContext, eveSig, aliceSig)
	require.NoError(t, err)
	_, authSetByte = chaincodeStub.PutStateArgsForCall(4)

	chaincodeStub.GetStateReturnsOnCall(7, authSetByte, nil)
	err = secretKeeper.LockSecret(transactionContext, bobSig, newSecret2)
	require.NoError(t, err)
	_, secretByte = chaincodeStub.PutStateArgsForCall(5)

	chaincodeStub.GetStateReturnsOnCall(8, authSetByte, nil)
	chaincodeStub.GetStateReturnsOnCall(9, secretByte, nil)
	secret, err = secretKeeper.RevealSecret(transactionContext, aliceSig)
	require.EqualError(t, err, "User are not allowed to perform this action.")
	require.Nil(t, secret)
}

func TestRollbackAttack(t *testing.T) {
	chaincodeStub := &fakes.ChaincodeStub{}
	transactionContext := &fakes.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	secretKeeper := SecretKeeper{}
	err := secretKeeper.InitSecretKeeper(transactionContext)
	require.NoError(t, err)

	_, authSetByte := chaincodeStub.PutStateArgsForCall(0)
	_, secretByte := chaincodeStub.PutStateArgsForCall(1)
	oldauthSetByte := authSetByte

	aliceSig := "Alice"
	bobSig := "Bob"
	eveSig := "Eve"
	newSecret := "NewSecret"
	newSecret2 := "SecretWithoutAlice"

	chaincodeStub.GetStateReturnsOnCall(0, authSetByte, nil)
	chaincodeStub.GetStateReturnsOnCall(1, authSetByte, nil)
	err = secretKeeper.AddUser(transactionContext, aliceSig, eveSig)
	require.NoError(t, err)
	_, authSetByte = chaincodeStub.PutStateArgsForCall(2)

	chaincodeStub.GetStateReturnsOnCall(2, authSetByte, nil)
	err = secretKeeper.LockSecret(transactionContext, eveSig, newSecret)
	require.NoError(t, err)
	_, secretByte = chaincodeStub.PutStateArgsForCall(3)

	chaincodeStub.GetStateReturnsOnCall(3, authSetByte, nil)
	chaincodeStub.GetStateReturnsOnCall(4, secretByte, nil)
	secret, err := secretKeeper.RevealSecret(transactionContext, aliceSig)
	require.NoError(t, err)
	require.EqualValues(t, secret.Value, newSecret)

	chaincodeStub.GetStateReturnsOnCall(5, authSetByte, nil)
	chaincodeStub.GetStateReturnsOnCall(6, authSetByte, nil)
	err = secretKeeper.RemoveUser(transactionContext, eveSig, aliceSig)
	require.NoError(t, err)

	_, authSetByte = chaincodeStub.PutStateArgsForCall(4)

	chaincodeStub.GetStateReturnsOnCall(7, authSetByte, nil)
	err = secretKeeper.LockSecret(transactionContext, bobSig, newSecret2)
	require.NoError(t, err)
	_, secretByte = chaincodeStub.PutStateArgsForCall(5)

	// Simulate rollback attack here
	chaincodeStub.GetStateReturnsOnCall(8, oldauthSetByte, nil)
	chaincodeStub.GetStateReturnsOnCall(9, secretByte, nil)
	secret, err = secretKeeper.RevealSecret(transactionContext, aliceSig)
	require.NoError(t, err)
	require.EqualValues(t, secret.Value, newSecret2)
}
