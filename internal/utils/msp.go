/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"

	"github.com/hyperledger/fabric/protoutil"
)

type IdentityEvaluatorInterface interface {
	EvaluateCreatorIdentity(creatorIdentityBytes []byte, ownerMSP string) error
}

type IdentityEvaluator struct {
}

// EvaluateCreatorIdentity check that two identities have the same msp id.
// This function requires marshalled msp.SerializedIdentity as inputs.
func (id *IdentityEvaluator) EvaluateCreatorIdentity(creatorIdentityBytes []byte, ownerMSP string) error {
	creatorMSP, err := ExtractMSPID(creatorIdentityBytes)
	if err != nil {
		return fmt.Errorf("error while deserialzing creator identity, err: %s", err)
	}

	if creatorMSP != ownerMSP {
		return fmt.Errorf("creator msp does not match owner msp")
	}

	return nil
}

func ExtractMSPID(serializedIdentityRaw []byte) (string, error) {
	sID, err := protoutil.UnmarshalSerializedIdentity(serializedIdentityRaw)
	if err != nil {
		return "", err
	}
	return sID.Mspid, nil
}
