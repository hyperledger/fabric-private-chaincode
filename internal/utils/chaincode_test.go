/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-chaincode-go/shimtest/mock"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChaincode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaincode Suite")
}

var _ = Describe("Chaincode", func() {

	var (
		stub              *shimtest.MockStub
		lscc              *mock.Chaincode
		expectedMrEnclave string
	)

	BeforeEach(func() {
		// create mock err
		stub = shimtest.NewMockStub("ercc", &mock.Chaincode{})
		stub.ChannelID = "mockChannel"

		lscc = &mock.Chaincode{}
		lifecycleStub := shimtest.NewMockStub("_lifecycle", lscc)
		stub.MockPeerChaincode("_lifecycle", lifecycleStub, stub.ChannelID)
	})

	When("chaincode definition exists", func() {
		BeforeEach(func() {
			expectedMrEnclave = "98aed61c91f258a37c68ed4943297695647ec7bbe6008cc111b0a12650ebeb91"

			// create mock lifecycle chaincode
			df := &lifecycle.QueryApprovedChaincodeDefinitionResult{
				Version: expectedMrEnclave,
			}
			dfBytes, err := proto.Marshal(df)
			Expect(err).ShouldNot(HaveOccurred())
			lscc.InvokeReturns(shim.Success(dfBytes))
		})

		It("should succeed", func() {
			df, err := utils.GetChaincodeDefinition("myFPCChaincode", stub)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(df).ShouldNot(BeNil())

			mrenclave, err := utils.ExtractMrEnclaveFromChaincodeDefinition(df)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(mrenclave).Should(Equal(expectedMrEnclave))
		})
	})

	When("chaincode definition exists", func() {
		It("should fail", func() {
			df, err := utils.GetChaincodeDefinition("myFPCChaincode", stub)
			Expect(err).Should(HaveOccurred())
			Expect(df).Should(BeNil())

			mrenclave, err := utils.ExtractMrEnclaveFromChaincodeDefinition(df)
			Expect(err).Should(HaveOccurred())
			Expect(mrenclave).Should(BeEmpty())
		})
	})

})
