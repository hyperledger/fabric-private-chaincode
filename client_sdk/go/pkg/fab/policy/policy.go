/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package policy

import (
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/common/policydsl"
)

// MajorityOfMembers returns an endorsement policy that requires the signatures
// by a majority of the given members.
//
// IMPORTANT: Note that this endorsement policy is not flexible enough to support
// membership changes and _might become *insecure* when membership increases.
// A more robust majority-based endorsement policy can be defined by not
// specifying an explicit policy and hence applying the channels default endorsement
// policy (which is "majority" in the default fabric network configurations).
// It is recommended to use this function in testing environments only where a
// static channel membership. For other deployments please use a
// "MAJORITY Endorsement"-based rule. For more details please see the official
// [Fabric documentation](https://hyperledger-fabric.readthedocs.io/en/latest/policies/policies.html).
//
// If the given members are empty, a reject all policy is returned.
func MajorityOfMembers(members []string) *cb.SignaturePolicyEnvelope {
	return majorityOf(members, msp.MSPRole_MEMBER)
}

func majorityOf(ids []string, role msp.MSPRole_MSPRoleType) *cb.SignaturePolicyEnvelope {
	if len(ids) < 1 {
		return policydsl.RejectAllPolicy
	}

	m := int32(len(ids)/2) + 1
	return policydsl.SignedByNOutOfGivenRole(m, role, ids)
}
