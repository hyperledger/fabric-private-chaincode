package utils

import (
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/protoutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chaincode utils", func() {

	Context("EvaluateCreatorIdentity", func() {

		var (
			eval *IdentityEvaluator
		)

		BeforeEach(func() {
			eval = &IdentityEvaluator{}
		})

		When("creatorIdentity is invalid", func() {
			It("should return an error", func() {
				sid := []byte("someGarbageBytes")
				err := eval.EvaluateCreatorIdentity(sid, "dummyMsp")
				Expect(err).Should(HaveOccurred())
			})
		})

		When("mspid mismatch", func() {
			It("should return an error", func() {
				sid := protoutil.MarshalOrPanic(&msp.SerializedIdentity{Mspid: "someMSP"})
				err := eval.EvaluateCreatorIdentity(sid, "dummyMsp")
				Expect(err).Should(HaveOccurred())
			})
		})

		When("mspid match", func() {
			It("should return no error", func() {
				sid := protoutil.MarshalOrPanic(&msp.SerializedIdentity{Mspid: "dummyMsp"})
				err := eval.EvaluateCreatorIdentity(sid, "dummyMsp")
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
