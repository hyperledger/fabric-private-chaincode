/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAttestation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Attestation Suite")
}
