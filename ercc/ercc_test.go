/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main_test

import (
	"fmt"
	"testing"

	main "github.com/hyperledger-labs/fabric-private-chaincode/ercc"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	enclavePK     = `MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEE9lPD9QkW9oxWlFvwABrmseYAVvoBvvmTt3jzV0sdASR2KDDQPvz8EcyqfomEOTwSz7E+mISktMxYqofRr+4Yw==`
	enclavePkHash = `qpEqqBaEkNz9bTO77QK8+CLbvaEN1NATs7ajRTzq70k=`
	quote         = `AgAAAG4NAAAEAAQAAAAAACVC+Q1jMSwdovbiGHbw44nMDb+CvAvF0FJF/38NWjOqAgIC/wEBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABwAAAAAAAAAHAAAAAAAAAJiu1hyR8lijfGjtSUMpdpVkfse75gCMwRGwoSZQ6+uRAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACD1xnnferKFHD2uvYqTXdDA8iZ22kCD5xw7h38CMfOngAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD3TvjLWa36sT/kCIRYXhtYoRQ61x2u48Q16bzoq8w6egAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAqAIAAFXdt6ObnofTxKVhK9Eafot/LsUGgr4546W34JUey7aqo9b6mpeP3X6W/DMIc1JbIXpFd5+mHWP+R7swgSEYNg+RUVbjkZ38nOJHzIl0E7Dgxs8X8iilH+hxpcPiYQphpIcBS5NUCmDn6Wsz/I+Dbpbt3e2G74WFPLHqDn+JHva5vtaYHd7cAfmPIhZZXXMCQJ8um5Jcer4L16VOugt8LEE0i3FqLb0khMYUHmEsqWuh1Fss5bNUuRDqotz6XTBq0uQ+nCfzv9ZsT2CDihsuQTzgU0BiZZuf06Aw9NQdywg+vTZoqyWw0Ca/jsAt+OpbQeQzDoH3HAvnaRvRByozHqKQ1Z83vVny2DQPVwWm6hxIEUCDVE2A/fkbo+UjR12fD8XWUw3xXfd6Dob9N2gBAAB1NRKH8uhAp94KvF/EF76xtBOYnlpAkbv4pYsmJfWkt0CtKtt/lvMQqkmZwSi8LQ93XBiAdVEKt255ycfFxcmAHPFPrjwHMb0/5wKNXa9vyBlgJ63tU/8U1JxujZ6QdS05xiQbKb+l2y6Nm++iw1Ba7BBJgQR+xDBud/VMjjLI3/nMlA9JTpVw9sSTsWdqHzA4bJm2P7fxkxL4wUYe6w+1uWGnT8XFwuJOfw1bUKZWlGZCOe8iLiPmDOmKUegpiLy0wY73gk+5bJhq1L8b4EXJMoSVoS4JgzYajh8oEBaUheiR4ze8sD9KuF0y+dfQklcMKdONyXMcI8QcZfj19iQy2FvXY8Ca0AoBkQMk4bn49e19ePChDUhrk7ynGGy5d9Wo8g3aNZLNWol5LuwCduTYv83xbHeKDkEsvk23m5NiXlVnDo6Pwu+32w57sX4K4CcojQZvJRYfFUuRCoN05TY0oJ0qvvZ1pAEAQAuBfOucbX6QZZ4qPcMR`
	mrenclave     = `mK7WHJHyWKN8aO1JQyl2lWR+x7vmAIzBEbChJlDr65E=`
)

func TestErcc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ercc Suite")
}

var _ = Describe("ERCC", func() {

	var (
		ercc *main.EnclaveRegistryCC
		stub *shimtest.MockStub
	)

	BeforeEach(func() {
		ercc = main.NewTestErcc()
		stub = shimtest.NewMockStub("ercc", ercc)
		stub.Decorations["test"] = []byte("AAA")
	})

	Context("instantiate chaincode", func() {
		It("should succeed", func() {
			res := stub.MockInit("1", [][]byte{})
			Expect(res.Status).To(Equal(int32(shim.OK)))
		})

	})

	Context("chaincode is instantiated", func() {
		BeforeEach(func() {
			res := stub.MockInit("1", [][]byte{})
			Expect(res.Status).To(Equal(int32(shim.OK)))
		})

		When("invoke register", func() {
			It("should succeed", func() {
				args := [][]byte{[]byte("registerEnclave"), []byte(enclavePK), []byte(quote), []byte(mrenclave)}
				res := stub.MockInvoke("1", args)
				fmt.Println(res.Message)
				Expect(res.Status).To(Equal(int32(shim.OK)))
			})

		})

		Context("invoke getAttestationReport", func() {
			When("report is registered", func() {
				BeforeEach(func() {
					args := [][]byte{[]byte("registerEnclave"), []byte(enclavePK), []byte(quote), []byte(mrenclave)}
					res := stub.MockInvoke("1", args)
					Expect(res.Status).To(Equal(int32(shim.OK)))
				})

				It("should succeed", func() {
					args := [][]byte{[]byte("getAttestationReport"), []byte(enclavePkHash)}
					res := stub.MockInvoke("2", args)
					Expect(res.Status).To(Equal(int32(shim.OK)))
				})
			})

			When("report is not registered", func() {
				It("should fail", func() {
					args := [][]byte{[]byte("getAttestationReport"), []byte(enclavePkHash)}
					res := stub.MockInvoke("2", args)
					Expect(res.Status).To(Equal(int32(shim.ERROR)))
					Expect(res.Message).To(Equal(fmt.Sprintf("EnclavePK does not exist: %s", enclavePkHash)))
				})
			})

		})
	})
})
