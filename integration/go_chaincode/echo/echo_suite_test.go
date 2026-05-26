/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package echo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/hyperledger-labs/fabric-smart-client/integration"
)

// fpcEchoPort is the base port for the FPC echo integration test suite.
const fpcEchoPort integration.TestPortRange = 20000

func TestEndToEnd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FPC Echo Suite")
}

func StartPort() int {
	return fpcEchoPort.StartPortForNode()
}
