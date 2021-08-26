/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils_test

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils/fakes"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/protoutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate counterfeiter -o fakes/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
//lint:ignore U1000 This is just used to generate fake
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

func TestChaincode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "utils test suite")
}

var _ = Describe("Chaincode utils", func() {

	var (
		stub              *fakes.ChaincodeStub
		expectedMrEnclave string
	)

	BeforeEach(func() {
		stub = &fakes.ChaincodeStub{}
		expectedMrEnclave = "98aed61c91f258a37c68ed4943297695647ec7bbe6008cc111b0a12650ebeb91"
	})

	Context("GetChaincodeDefinition", func() {
		When("committed chaincode definition exists at _lifecycle", func() {
			BeforeEach(func() {
				// register chaincode definition at _lifecycle
				df := &lifecycle.QueryChaincodeDefinitionResult{
					Sequence: 666,
					Version:  "someVersion",
				}
				dfBytes, err := protoutil.Marshal(df)
				Expect(err).ShouldNot(HaveOccurred())
				stub.InvokeChaincodeReturns(shim.Success(dfBytes))
			})

			It("should return chaincode definition", func() {
				df, err := utils.GetChaincodeDefinition("myFPCChaincode", stub)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(df).ShouldNot(BeNil())
			})
		})

		When("committed chaincode definition does not exist at _lifecycle", func() {
			It("should return error", func() {
				stub.InvokeChaincodeReturns(peer.Response{Payload: nil, Status: shim.OK})
				df, err := utils.GetChaincodeDefinition("myFPCChaincode", stub)
				Expect(err).Should(MatchError(fmt.Errorf("no chaincode definition found for chaincode='myFPCChaincode'")))
				Expect(df).Should(BeNil())
			})
		})

		When("chaincode definition bytes are not valid", func() {
			BeforeEach(func() {
				// register chaincode definition at _lifecycle
				dfBytes := []byte{0x00, 0x12}
				stub.InvokeChaincodeReturns(shim.Success(dfBytes))
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

				dfBytes, err := protoutil.Marshal(df)
				Expect(err).ShouldNot(HaveOccurred())
				stub.InvokeChaincodeReturns(shim.Success(dfBytes))
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
					Sequence: 666,
					Version:  "",
				}

				dfBytes, err := protoutil.Marshal(df)
				Expect(err).ShouldNot(HaveOccurred())
				stub.InvokeChaincodeReturns(shim.Success(dfBytes))
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

				dfBytes, err := protoutil.Marshal(df)
				Expect(err).ShouldNot(HaveOccurred())
				stub.InvokeChaincodeReturns(shim.Success(dfBytes))
			})

			It("should return error", func() {
				df, err := utils.GetMrEnclave("myFPCChaincode", stub)
				Expect(err).Should(HaveOccurred())
				Expect(df).Should(BeEmpty())
			})
		})
	})
})
