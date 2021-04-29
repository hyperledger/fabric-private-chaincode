/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils_test

import (
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/common/policydsl"
	"github.com/hyperledger/fabric/core/config/configtest"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/msp/mgmt"
	"github.com/hyperledger/fabric/protoutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChaincode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaincode Suite")
}

var _ = Describe("Chaincode", func() {

	var (
		cryptoProvider bccsp.BCCSP
		serializedId   []byte
		pe             utils.PolicyEvaluatorInterface
	)

	BeforeSuite(func() {
		mspDir := configtest.GetDevMspDir()
		testConf, err := msp.GetLocalMspConfig(mspDir, nil, "SampleOrg")
		Expect(err).ShouldNot(HaveOccurred())

		cryptoProvider = factory.GetDefault()
		err = mgmt.GetLocalMSP(cryptoProvider).Setup(testConf)
		Expect(err).ShouldNot(HaveOccurred())

		pe = utils.NewPolicyEvaluator()
	})

	BeforeEach(func() {
		i, err := mgmt.GetLocalMSP(cryptoProvider).GetDefaultSigningIdentity()
		Expect(err).ShouldNot(HaveOccurred())

		serializedId, err = i.Serialize()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("IsValidEndorserIdentity", func() {

		When("identity satisfies endorsement policy", func() {
			It("should succeed", func() {
				p, err := policydsl.FromString("OR('SampleOrg.member', 'AnotherOtherOrg.member')")
				Expect(err).ShouldNot(HaveOccurred())
				pp := marshalApplicationPolicy(p, "")

				df := &lifecycle.QueryChaincodeDefinitionResult{
					ValidationParameter: pp,
				}
				err = pe.EvaluateIdentity(df.ValidationParameter, serializedId)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("identity does not satisfy endorsement policy", func() {
			It("should return error", func() {
				p, err := policydsl.FromString("OR('SomeOtherOrg.member', 'AnotherOtherOrg.member')")
				Expect(err).ShouldNot(HaveOccurred())
				pp := marshalApplicationPolicy(p, "")

				df := &lifecycle.QueryChaincodeDefinitionResult{
					ValidationParameter: pp,
				}
				err = pe.EvaluateIdentity(df.ValidationParameter, serializedId)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
})

func marshalApplicationPolicy(signaturePolicy *cb.SignaturePolicyEnvelope, channelConfigPolicy string) []byte {
	if signaturePolicy == nil && channelConfigPolicy == "" {
		panic("inputs empty")
	}

	if signaturePolicy != nil && channelConfigPolicy != "" {
		panic("cannot specify both signature policy and channel config policy")
	}

	var applicationPolicy *pb.ApplicationPolicy
	if signaturePolicy != nil {
		applicationPolicy = &pb.ApplicationPolicy{
			Type: &pb.ApplicationPolicy_SignaturePolicy{
				SignaturePolicy: signaturePolicy,
			},
		}
	}

	if channelConfigPolicy != "" {
		applicationPolicy = &pb.ApplicationPolicy{
			Type: &pb.ApplicationPolicy_ChannelConfigPolicyReference{
				ChannelConfigPolicyReference: channelConfigPolicy,
			},
		}
	}

	return protoutil.MarshalOrPanic(applicationPolicy)
}
