/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
)

type IdentityEvaluatorInterface interface {
	EvaluateCreatorIdentity(creatorIdentityBytes []byte, ownerIdentityBytes []byte) error
}

type IdentityEvaluator struct {
}

// EvaluateCreatorIdentity check that two identities have the same msp id.
// This function requires marshalled msp.SerializedIdentity as inputs.
func (id *IdentityEvaluator) EvaluateCreatorIdentity(creatorIdentityBytes []byte, ownerIdentityBytes []byte) error {
	creatorMSP, err := extractMSPID(creatorIdentityBytes)
	if err != nil {
		return fmt.Errorf("error while deserialzing creator identity, err: %s", err)
	}

	ownerMSP, err := extractMSPID(ownerIdentityBytes)
	if err != nil {
		return fmt.Errorf("error while deserialzing owner identity, err: %s", err)
	}

	if creatorMSP != ownerMSP {
		return fmt.Errorf("creator msp does not match owner msp")
	}

	return nil
}

func extractMSPID(serializedIdentityRaw []byte) (string, error) {
	sID := &msp.SerializedIdentity{}
	err := proto.Unmarshal(serializedIdentityRaw, sID)
	if err != nil {
		return "", err
	}
	return sID.Mspid, nil
}
