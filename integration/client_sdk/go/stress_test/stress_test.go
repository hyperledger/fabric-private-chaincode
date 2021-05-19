/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package stress_test_test

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"testing"

	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"github.com/hyperledger/fabric-private-chaincode/integration/client_sdk/go/utils"
	testutils "github.com/hyperledger/fabric-private-chaincode/integration/client_sdk/go/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStress(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaincode Suite")
}

var (
	network      *gateway.Network
	echoContract fpc.Contract
)

var _ = BeforeSuite(func() {
	ccID := "kv-test"
	ccPath := filepath.Join(utils.FPCPath, "samples", "chaincode", ccID, "_build", "lib")

	// setup stress test chaincode(s) (install, approve, commit)
	err := testutils.Setup(ccID, ccPath, true)
	Expect(err).ShouldNot(HaveOccurred())

	// get network
	network, err = testutils.SetupNetwork("mychannel")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(network).ShouldNot(BeNil())

	// Get FPC Contract
	echoContract = fpc.GetContract(network, ccID)
	Expect(echoContract).ShouldNot(BeNil())
})

var _ = Describe("Stress tests", func() {
	Context("Different payload sizes", func() {
		When("submitting less than MAX_ARGUMENT_SIZE", func() {
			It("should succeed", func() {
				sizes := []int{1, 10, 100, 1000, 10000}
				for _, size := range sizes {
					fmt.Printf("Size = %d\n", size)

					payload := make([]byte, size)
					n, err := rand.Read(payload)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(n).Should(Equal(size))

					key := "some-key"
					value := base64.StdEncoding.EncodeToString(payload)

					res, err := echoContract.SubmitTransaction("put_state", key, value)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal([]byte("OK")))

					res, err = echoContract.EvaluateTransaction("get_state", key)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal([]byte(value)))
				}
			})
		})
	})

	// TODO more stress tests added here
})
