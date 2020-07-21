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
		stub      *shimtest.MockStub
		lscc      *mock.Chaincode
		mrenclave string
	)

	BeforeEach(func() {

		mrenclave = "helloMrEnclave"

		// create mock err
		stub = shimtest.NewMockStub("ercc", &mock.Chaincode{})
		stub.ChannelID = "mockChannel"

		// create mock lifecycle chaincode
		df := &lifecycle.QueryApprovedChaincodeDefinitionResult{
			Version: mrenclave,
		}
		dfBytes, err := proto.Marshal(df)
		Expect(err).ShouldNot(HaveOccurred())

		lscc = &mock.Chaincode{}
		lscc.InvokeReturns(shim.Success(dfBytes))
		lifecycleStub := shimtest.NewMockStub("_lifecycle", lscc)

		stub.MockPeerChaincode("_lifecycle", lifecycleStub, stub.ChannelID)
	})

	It("should return a chaincode definition", func() {
		df, err := utils.ExtractMrEnclaveFromChaincodeDefinition("myFPCChaincode", stub)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(df).Should(Equal(mrenclave))
	})

})
