/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package integration_test

import (
	"fmt"
	"path/filepath"
	"testing"

	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	testutils "github.com/hyperledger/fabric-private-chaincode/integration/client_sdk/go/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoClientSDK(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaincode Suite")
}

var (
	network        *gateway.Network
	contract       fpc.Contract
	auctionCounter int
	auctionName    string
	numClients     int
)

var _ = BeforeSuite(func() {
	ccID := "auction_test"
	ccPath := filepath.Join(testutils.FPCPath, "samples", "chaincode", "auction", "_build", "lib")

	// setup auction chaincode (install, approve, commit)
	initEnclave := true
	err := testutils.Setup(ccID, ccPath, initEnclave)
	Expect(err).ShouldNot(HaveOccurred())

	// setup echo chaincode (install, approve, commit)
	err = testutils.Setup("echo_test", filepath.Join(testutils.FPCPath, "samples", "chaincode", "echo", "_build", "lib"), false)
	Expect(err).ShouldNot(HaveOccurred())

	// get network
	network, err = testutils.SetupNetwork("mychannel")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(network).ShouldNot(BeNil())

	// Get FPC Contract (auction)
	contract = fpc.GetContract(network, ccID)
	Expect(contract).ShouldNot(BeNil())
})

var _ = Describe("Go Client SDK Test", func() {

	BeforeEach(func() {
		res, err := contract.SubmitTransaction("init", "MyAuctionHouse")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(res).Should(Equal([]byte("OK")))
	})

	Context("Scenario 1", func() {
		When("close non existing auction.", func() {
			It("should return AUCTION_NOT_EXISTING", func() {
				res, err := contract.SubmitTransaction("close", "MyAuction")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("AUCTION_NOT_EXISTING")))
			})
		})

		When("evaluate non existing auction.", func() {
			It("should return AUCTION_NOT_EXISTING", func() {
				res, err := contract.SubmitTransaction("eval", "MyAuction0")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("AUCTION_NOT_EXISTING")))
			})
		})
	})

	Context("Scenario 2", func() {
		BeforeEach(func() {
			auctionCounter++
			auctionName = fmt.Sprintf("MyAuction%d", auctionCounter)

			res, err := contract.SubmitTransaction("create", auctionName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).Should(Equal([]byte("OK")))
		})

		When("create two equivalent bids", func() {
			It("should return DRAW", func() {
				res, err := contract.SubmitTransaction("submit", auctionName, "JohnnyCash0", "2")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("OK")))

				res, err = contract.SubmitTransaction("submit", auctionName, "JohnnyCash1", "2")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("OK")))

				res, err = contract.SubmitTransaction("close", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("OK")))

				res, err = contract.SubmitTransaction("submit", auctionName, "JohnnyCash2", "2")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("AUCTION_ALREADY_CLOSED")))

				res, err = contract.SubmitTransaction("eval", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("DRAW")))
			})
		})
	})

	Context("Scenario 3", func() {
		BeforeEach(func() {
			auctionCounter++
			auctionName = fmt.Sprintf("MyAuction%d", auctionCounter)
			numClients = 10

			res, err := contract.SubmitTransaction("create", auctionName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).Should(Equal([]byte("OK")))
		})

		When("submit unique bids", func() {
			It("should print out the auction result", func() {
				for i := 0; i <= numClients; i++ {
					res, err := contract.SubmitTransaction("submit", auctionName, fmt.Sprintf("JohnnyCash%d", i), fmt.Sprintf("%d", i))
					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal([]byte("OK")))
				}

				res, err := contract.SubmitTransaction("close", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("OK")))

				res, err = contract.SubmitTransaction("eval", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte(fmt.Sprintf("{\"bidder\":\"JohnnyCash%d\",\"value\":%d}", numClients, numClients))))
			})
		})
	})

	Context("Scenario 4", func() {
		BeforeEach(func() {
			auctionCounter++
			auctionName = fmt.Sprintf("MyAuction%d", auctionCounter)

			res, err := contract.SubmitTransaction("create", auctionName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).Should(Equal([]byte("OK")))
		})

		When("create a duplicate auction", func() {
			It("should return AUCTION_ALREADY_EXISTING", func() {
				res, err := contract.SubmitTransaction("create", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("AUCTION_ALREADY_EXISTING")))
			})
		})

		When("close auction twice", func() {
			It("should return AUCTION_ALREADY_CLOSED", func() {
				res, err := contract.SubmitTransaction("close", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("OK")))

				res, err = contract.SubmitTransaction("close", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("AUCTION_ALREADY_CLOSED")))
			})
		})

		When("close auction via query (evalaute transaction)", func() {
			It("should return OK when closing again", func() {
				res, err := contract.EvaluateTransaction("close", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("OK")))

				res, err = contract.SubmitTransaction("close", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("OK")))
			})
		})

		When("no bids submitted", func() {
			It("should return AUCTION_ALREADY_CLOSED", func() {
				res, err := contract.SubmitTransaction("close", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("OK")))

				res, err = contract.SubmitTransaction("eval", auctionName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).Should(Equal([]byte("NO_BIDS")))
			})
		})
	})

	Context("Base SDK Tests with Auction and Echo", func() {
		When("invoking/querying a non-existing fpc chaincode", func() {
			It("should return error", func() {
				contract := fpc.GetContract(network, "not_installed")
				Expect(contract).ShouldNot(BeNil())

				res, err := contract.EvaluateTransaction("do", "something")
				Expect(err).Should(HaveOccurred())
				Expect(res).Should(BeNil())

				res, err = contract.SubmitTransaction("do", "something", "else")
				Expect(err).Should(HaveOccurred())
				Expect(res).Should(BeNil())
			})
		})

		When("invoking/querying a non-registered fpc chaincode (echo test)", func() {
			It("should return error", func() {
				contract := fpc.GetContract(network, "echo_test")
				Expect(contract).ShouldNot(BeNil())

				res, err := contract.EvaluateTransaction("do", "something")
				Expect(err).Should(HaveOccurred())
				Expect(res).Should(BeNil())

				res, err = contract.SubmitTransaction("do", "something", "else")
				Expect(err).Should(HaveOccurred())
				Expect(res).Should(BeNil())
			})
		})
	})
})
