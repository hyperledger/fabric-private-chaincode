/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"

	//lint:ignore SA1019 old protos required for fpc.pb.go and test compatibility
	protoV1 "github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
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
		return fmt.Errorf("error while deserialzing creator identity, err: %w", err)
	}

	if creatorMSP != ownerMSP {
		return fmt.Errorf("creator msp does not match owner msp")
	}

	return nil
}

func ExtractMSPID(serializedIdentityRaw []byte) (string, error) {
	sID := &msp.SerializedIdentity{}
	if err := protoV1.Unmarshal(serializedIdentityRaw, sID); err != nil {
		return "", err
	}
	return sID.Mspid, nil
}
