/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation_test

import (
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Verifier", func() {

	When("invoke VerifyAttestationReport with incorrect Report", func() {
		It("should return Intel pub key", func() {
			verifier := &attestation.VerifierImpl{}
			res, err := verifier.VerifyAttestationReport("someKey", attestation.IASAttestationReport{})
			Expect(err).To(MatchError("provided cert not PEM formatted"))
			Expect(res).To(BeFalse())
		})
	})

	// TODO work on testing
})
