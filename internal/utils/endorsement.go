package utils

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/msp/mgmt"
)

func NewPolicyEvaluator() *PolicyEvaluator {
	manager := mgmt.NewDeserializersManager(factory.GetDefault())
	return &PolicyEvaluator{
		IdentityDeserializer: manager.GetLocalDeserializer(),
	}
}

type PolicyEvaluator struct {
	msp.IdentityDeserializer
}

func (id *PolicyEvaluator) EvaluateIdentity(policyBytes []byte, identityBytes []byte) error {
	identity, err := id.IdentityDeserializer.DeserializeIdentity(identityBytes)
	if err != nil {
		return fmt.Errorf("error while deserialzing identity, err: %s", err)
	}

	pp := cauthdsl.NewPolicyProvider(id.IdentityDeserializer)
	policy, _, err := pp.NewPolicy(policyBytes)
	if err != nil {
		return err
	}

	return policy.EvaluateIdentities([]msp.Identity{identity})
}

func IsValidEndorserIdentity(identityBytes []byte, ccDef *lifecycle.QueryApprovedChaincodeDefinitionResult) error {
	evaluator := NewPolicyEvaluator()
	return evaluator.EvaluateIdentity(ccDef.ValidationParameter, identityBytes)
}
