/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils_test

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-chaincode-go/shimtest/mock"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chaincode utils", func() {

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

		expectedMrEnclave = "98aed61c91f258a37c68ed4943297695647ec7bbe6008cc111b0a12650ebeb91"
	})

	Context("GetChaincodeDefinition", func() {
		When("approved chaincode definition exists at _lifecycle", func() {
			BeforeEach(func() {
				// register chaincode definition at _lifecycle
				df := &lifecycle.QueryChaincodeDefinitionResult{}
				dfBytes, err := proto.Marshal(df)
				Expect(err).ShouldNot(HaveOccurred())
				lscc.InvokeReturns(shim.Success(dfBytes))
			})

			It("should return chaincode definition", func() {
				df, err := utils.GetChaincodeDefinition("myFPCChaincode", stub)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(df).ShouldNot(BeNil())
			})
		})

		When("approved chaincode definition does not exist at _lifecycle", func() {
			It("should return error", func() {
				df, err := utils.GetChaincodeDefinition("myFPCChaincode", stub)
				Expect(err).Should(MatchError(fmt.Errorf("no chaincode definition found for chaincode='myFPCChaincode'")))
				Expect(df).Should(BeNil())
			})
		})

		When("chaincode definition bytes are not valid", func() {
			BeforeEach(func() {
				// register chaincode definition at _lifecycle
				dfBytes := []byte{0x00, 0x12}
				lscc.InvokeReturns(shim.Success(dfBytes))
			})

			It("should return QueryChaincodeDefinitionResult object", func() {
				df, err := utils.GetChaincodeDefinition("myFPCChaincode", stub)
				Expect(err).Should(HaveOccurred())
				Expect(df).Should(BeNil())
			})
		})
	})

	Context("GetMrEnclave", func() {
		When("mrenclave is valid", func() {
			BeforeEach(func() {
				df := &lifecycle.QueryChaincodeDefinitionResult{
					Version: expectedMrEnclave,
				}

				dfBytes, err := proto.Marshal(df)
				Expect(err).ShouldNot(HaveOccurred())
				lscc.InvokeReturns(shim.Success(dfBytes))
			})

			It("should return error", func() {
				mrenclave, err := utils.GetMrEnclave("myFPCChaincode", stub)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(mrenclave).Should(Equal(expectedMrEnclave))
			})
		})

		When("mrenclave is empty", func() {
			BeforeEach(func() {
				df := &lifecycle.QueryChaincodeDefinitionResult{
					Version: "",
				}

				dfBytes, err := proto.Marshal(df)
				Expect(err).ShouldNot(HaveOccurred())
				lscc.InvokeReturns(shim.Success(dfBytes))
			})

			It("should return error", func() {
				df, err := utils.GetMrEnclave("myFPCChaincode", stub)
				Expect(err).Should(MatchError(fmt.Errorf("mrenclave has wrong length! expteced %d but got %d", utils.MrEnclaveLength, 0)))
				Expect(df).Should(BeEmpty())
			})
		})

		When("mrenclave is not hexstring", func() {
			BeforeEach(func() {
				df := &lifecycle.QueryChaincodeDefinitionResult{
					Version: "mK7WHJHyWKN8aO1JQyl2lWR+x7vmAIzBEbChJlDr65E=",
				}

				dfBytes, err := proto.Marshal(df)
				Expect(err).ShouldNot(HaveOccurred())
				lscc.InvokeReturns(shim.Success(dfBytes))
			})

			It("should return error", func() {
				df, err := utils.GetMrEnclave("myFPCChaincode", stub)
				Expect(err).Should(HaveOccurred())
				Expect(df).Should(BeEmpty())
			})
		})
	})
})
