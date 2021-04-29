/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ccpackager_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/fab/ccpackager"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoClientSDK(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client SDK Test Suite")
}

var (
	fpcPath   string
	ccId      string
	ccPath    string
	mrenclave string
)

var _ = BeforeSuite(func() {
	fpcPath = os.Getenv("FPC_PATH")
	if fpcPath == "" {
		panic("FPC_PATH not set")
	}
})

var _ = Describe("Go Client SDK Test", func() {
	BeforeEach(func() {
		ccPath = filepath.Join(fpcPath, "samples", "chaincode", "auction", "_build", "lib")
		var err error
		mrenclave, err = ccpackager.ReadMrenclave(ccPath)
		Expect(err).ShouldNot(HaveOccurred())

		ccId = "auction"
	})

	Context("fpc-c", func() {
		When("SGX_MODE is not set", func() {
			It("should return an error", func() {
				desc := &ccpackager.Descriptor{
					Path:  ccPath,
					Type:  ccpackager.ChaincodeType,
					Label: ccId,
				}
				_, err := ccpackager.NewCCPackage(desc)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("enclave file does not exist", func() {
			It("should return an error", func() {
				desc := &ccpackager.Descriptor{
					Path:    "a/path/to/somewhere",
					Type:    ccpackager.ChaincodeType,
					Label:   ccId,
					SGXMode: sgx.SGXModeSimType,
				}
				_, err := ccpackager.NewCCPackage(desc)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("chaincode type is not set fpc-c", func() {
			It("should return an error", func() {
				desc := &ccpackager.Descriptor{
					Path:    ccPath,
					Type:    "go",
					Label:   ccId,
					SGXMode: sgx.SGXModeSimType,
				}
				_, err := ccpackager.NewCCPackage(desc)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("everything is set", func() {
			It("should work fine", func() {
				desc := &ccpackager.Descriptor{
					Path:    ccPath,
					Type:    ccpackager.ChaincodeType,
					Label:   ccId,
					SGXMode: sgx.SGXModeSimType,
				}
				ccPkg, err := ccpackager.NewCCPackage(desc)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ccPkg).ShouldNot(BeNil())
			})
		})
	})

	Context("caas", func() {
		When("CaaSEndpoint not set", func() {
			It("should return an error", func() {
				desc := &ccpackager.Descriptor{
					Type:  ccpackager.CaaSType,
					Label: ccId,
				}
				_, err := ccpackager.NewCCPackage(desc)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("CaaSEndpoint has invalid port", func() {
			It("should return an error", func() {
				desc := &ccpackager.Descriptor{
					Type:         ccpackager.CaaSType,
					CaaSEndpoint: "mychaincode.peer.example:invalidPort",
					Label:        ccId,
				}
				_, err := ccpackager.NewCCPackage(desc)
				Expect(err).Should(HaveOccurred())

			})
		})

		When("CaaSEndpoint no port", func() {
			It("should return an error", func() {
				// no port
				desc := &ccpackager.Descriptor{
					Type:         ccpackager.CaaSType,
					CaaSEndpoint: "mychaincode.peer.example",
					Label:        ccId,
				}
				_, err := ccpackager.NewCCPackage(desc)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("Label not set", func() {
			It("should return an error", func() {
				desc := &ccpackager.Descriptor{
					Type:         ccpackager.CaaSType,
					CaaSEndpoint: "mychaincode.peer.example:8123",
				}
				_, err := ccpackager.NewCCPackage(desc)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("Label has wrong format", func() {
			It("should return an error", func() {
				desc := &ccpackager.Descriptor{
					Type:         ccpackager.CaaSType,
					CaaSEndpoint: "mychaincode.peer.example:8123",
					Label:        "_invalid_Label",
				}
				_, err := ccpackager.NewCCPackage(desc)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("all good", func() {
			It("should work", func() {
				desc := &ccpackager.Descriptor{
					Type:         ccpackager.CaaSType,
					CaaSEndpoint: "mychaincode.peer.example:8123",
					Label:        ccId,
				}
				payload, err := ccpackager.NewCCPackage(desc)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(payload).ShouldNot(BeNil())

				err = ioutil.WriteFile("/tmp/pack.tar.gz", payload, 0644)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
