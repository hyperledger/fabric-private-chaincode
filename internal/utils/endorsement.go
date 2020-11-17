/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	pmsp "github.com/hyperledger/fabric-protos-go/msp"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

type PolicyEvaluatorInterface interface {
	EvaluateIdentity(policyBytes []byte, identityBytes []byte) error
}

type PolicyEvaluator struct {
}

func NewPolicyEvaluator() PolicyEvaluatorInterface {
	return &PolicyEvaluator{}
}

// EvaluateIdentity checks that the provided identity is a valid endorser as specified in the endorsement policy
// This function requires a marshalled pb.ApplicationPolicy and a marshalled msp.SerializedIdentity as input.
func (id *PolicyEvaluator) EvaluateIdentity(policyBytes []byte, identityBytes []byte) error {
	aMsp, err := ExtractMSPID(identityBytes)
	if err != nil {
		return fmt.Errorf("error while deserialzing creator identity, err: %s", err)
	}

	sp, ref, err := unmarshalApplicationPolicy(policyBytes)
	if err != nil {
		return fmt.Errorf("cannot convert application policy to signature policy, err: %s", err)
	}

	if ref != "" {
		// note that this is tricky issue. if the chaincode definition contains a reference to an EP such as
		// the channels default EP, the reference must be resolved with the help of the peer. since we are
		// using this code inside a chaincode (enclave registry) it cannot be resolved easily.
		// TODO find a way to resolve this reference in chaincode environment
		//
		// Current work around: explicit endorsement policies for fpc chaincode
		return fmt.Errorf("endorsement policy reference is provided, cannot parse, err: %s", err)
	}

	endorserMSPs, err := getMSPIDsFromSP(sp)
	if err != nil {
		return fmt.Errorf("cannot extract msp ids from signature policy, err: %s", err)
	}

	// TODO check that role matches

	if _, ok := endorserMSPs[aMsp]; !ok {
		return fmt.Errorf("identity is not a valid endorser")
	}

	return nil
}

func unmarshalApplicationPolicy(policyBytes []byte) (*common.SignaturePolicyEnvelope, string, error) {
	applicationPolicy := &pb.ApplicationPolicy{}
	err := proto.Unmarshal(policyBytes, applicationPolicy)
	if err != nil {
		return nil, "", errors.WithMessage(err, "failed to unmarshal application policy")
	}

	switch policy := applicationPolicy.Type.(type) {
	case *pb.ApplicationPolicy_SignaturePolicy:
		return policy.SignaturePolicy, "", nil
	case *pb.ApplicationPolicy_ChannelConfigPolicyReference:
		return nil, policy.ChannelConfigPolicyReference, nil
	default:
		return nil, "", errors.Errorf("unsupported policy type %T", policy)
	}
}

func getMSPIDsFromSP(sp *common.SignaturePolicyEnvelope) (map[string]string, error) {
	m := make(map[string]string)
	for _, identity := range sp.Identities {
		if identity.PrincipalClassification == pmsp.MSPPrincipal_ROLE {
			msprole := &pmsp.MSPRole{}
			err := proto.Unmarshal(identity.Principal, msprole)
			if err != nil {
				return nil, err
			}
			m[msprole.GetMspIdentifier()] = msprole.GetRole().String()
		}
	}
	return m, nil
}
