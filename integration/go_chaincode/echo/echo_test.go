/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package echo_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/hyperledger-labs/fabric-smart-client/integration"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/common"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fsc"
	fabricsdk "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/sdk/dig"
	fpcnwo "github.com/hyperledger/fabric-private-chaincode/extension/fsc/integration/nwo/fabric/fpc"
	"github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode/echo"
	"github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode/echo/views"
)

var _ = Describe("EndToEnd", func() {
	Describe("Echo FPC With LibP2P", func() {
		s := NewTestSuite(fsc.LibP2P, integration.NoReplication)
		BeforeEach(s.Setup)
		AfterEach(s.TearDown)
		It("succeeded", s.TestSucceeded)
	})

	Describe("Echo FPC With Websockets", func() {
		s := NewTestSuite(fsc.WebSocket, integration.NoReplication)
		BeforeEach(s.Setup)
		AfterEach(s.TearDown)
		It("succeeded", s.TestSucceeded)
	})
})

type TestSuite struct {
	*integration.TestSuite
}

func NewTestSuite(commType fsc.P2PCommunicationType, nodeOpts *integration.ReplicationOptions) *TestSuite {
	return &TestSuite{integration.NewTestSuite(func() (*integration.Infrastructure, error) {
		// RegisterPlatformFactory must happen BEFORE Generate so that initNWO
		// picks up the FPC factory when it instantiates the fabric platform.
		ii, err := integration.New(StartPort(), "", echo.Topology(&fabricsdk.SDK{}, commType, nodeOpts)...)
		if err != nil {
			return nil, err
		}
		ii.EnableRaceDetector()
		ii.RegisterPlatformFactory(fpcnwo.NewPlatformFactory())
		ii.Generate()
		return ii, nil
	})}
}

func (s *TestSuite) TestSucceeded() {
	provisionedEnclavesBoxed, err := s.II.Client("alice").CallView(
		"ListProvisionedEnclaves",
		common.JSONMarshall(&views.ListProvisionedEnclaves{
			CID: "echo",
		}),
	)
	Expect(err).ToNot(HaveOccurred())
	var provisionedEnclaves []string
	common.JSONUnmarshal(provisionedEnclavesBoxed.([]byte), &provisionedEnclaves)
	Expect(len(provisionedEnclaves)).To(BeEquivalentTo(1))

	resBoxed, err := s.II.Client("alice").CallView(
		"Echo",
		common.JSONMarshall(&views.Echo{
			Function: "myFunction",
			Args:     []string{"arg1", "arg2", "arg3"},
		}),
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(resBoxed).To(BeEquivalentTo("myFunction"))
}
