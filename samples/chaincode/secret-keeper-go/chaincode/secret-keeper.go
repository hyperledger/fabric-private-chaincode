/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const OK = "OK"
const AUTH_LIST_KEY = "AUTH_LIST_KEY"
const SECRET_KEY = "SECRET_KEY"

type SecretKeeper struct {
	contractapi.Contract
}

type AuthSet struct {
	Pubkey map[string]struct{}
}

type Secret struct {
	Value string `json:Value`
}

func (t *SecretKeeper) InitSecretKeeper(ctx contractapi.TransactionContextInterface) error {
	// init authSet
	pubkeyset := make(map[string]struct{})
	pubkeyset["Alice"] = struct{}{}
	pubkeyset["Bob"] = struct{}{}
	authSet := AuthSet{
		Pubkey: pubkeyset,
	}

	authSetJson, err := json.Marshal(authSet)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(AUTH_LIST_KEY, authSetJson)
	if err != nil {
		return fmt.Errorf("failed to put %s to world state. %v", AUTH_LIST_KEY, err)
	}

	// init secret
	secret := Secret{
		Value: "DefaultSecret",
	}

	secretJson, err := json.Marshal(secret)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(SECRET_KEY, secretJson)
	if err != nil {
		return fmt.Errorf("failed to put %s to world state. %v", SECRET_KEY, err)
	}

	return nil
}

func (t *SecretKeeper) AddUser(ctx contractapi.TransactionContextInterface, sig string, pubkey string) error {
	// check if the user allow to update authSet
	valid, err := VerifySig(ctx, sig)
	if err != nil {
		return err
	}
	if valid != true {
		return fmt.Errorf("User are not allowed to perform this action.")
	}

	// update the value
	authSet, err := GetAuthList(ctx)
	authSet.Pubkey[pubkey] = struct{}{}

	authSetJson, err := json.Marshal(authSet)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(AUTH_LIST_KEY, authSetJson)
	if err != nil {
		return fmt.Errorf("failed to put %s to world state. %v", AUTH_LIST_KEY, err)
	}

	return nil
}

func (t *SecretKeeper) RemoveUser(ctx contractapi.TransactionContextInterface, sig string, pubkey string) error {
	// check if the user allow to update authSet
	valid, err := VerifySig(ctx, sig)
	if err != nil {
		return err
	}
	if valid != true {
		return fmt.Errorf("User are not allowed to perform this action.")
	}

	// update the value
	authSet, err := GetAuthList(ctx)
	delete(authSet.Pubkey, pubkey)

	authSetJson, err := json.Marshal(authSet)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(AUTH_LIST_KEY, authSetJson)
	if err != nil {
		return fmt.Errorf("failed to put %s to world state. %v", AUTH_LIST_KEY, err)
	}

	return nil
}

func (t *SecretKeeper) LockSecret(ctx contractapi.TransactionContextInterface, sig string, value string) error {
	// check if the user allow to update secret
	valid, err := VerifySig(ctx, sig)
	if err != nil {
		return err
	}
	if valid != true {
		return fmt.Errorf("User are not allowed to perform this action.")
	}

	// update the value
	newSecret := Secret{
		Value: value,
	}

	newSecretJson, err := json.Marshal(newSecret)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(SECRET_KEY, newSecretJson)
	if err != nil {
		return fmt.Errorf("failed to put %s to world state. %v", SECRET_KEY, err)
	}

	return nil
}

func (t *SecretKeeper) RevealSecret(ctx contractapi.TransactionContextInterface, sig string) (*Secret, error) {
	// check if the user allow to view the secret.
	valid, err := VerifySig(ctx, sig)
	if err != nil {
		return nil, err
	}
	if valid != true {
		return nil, fmt.Errorf("User are not allowed to perform this action.")
	}

	// reveal secret
	secretJson, err := ctx.GetStub().GetState(SECRET_KEY)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if secretJson == nil {
		return nil, fmt.Errorf("the asset %s does not exist", SECRET_KEY)
	}
	var secret Secret
	err = json.Unmarshal(secretJson, &secret)
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

func GetAuthList(ctx contractapi.TransactionContextInterface) (*AuthSet, error) {
	authSetJson, err := ctx.GetStub().GetState(AUTH_LIST_KEY)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if authSetJson == nil {
		return nil, fmt.Errorf("the asset %s does not exist", AUTH_LIST_KEY)
	}

	var authSet AuthSet
	err = json.Unmarshal(authSetJson, &authSet)
	if err != nil {
		return nil, err
	}
	return &authSet, nil
}

func VerifySig(ctx contractapi.TransactionContextInterface, sig string) (bool, error) {
	authSet, err := GetAuthList(ctx)
	if err != nil {
		return false, err
	}

	if _, exist := authSet.Pubkey[sig]; exist {
		return true, nil
	}

	return false, nil
}
