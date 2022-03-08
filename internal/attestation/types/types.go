/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package types

type ConvertFunction func(attestationBytes []byte) (evidenceBytes []byte, err error)

type Converter struct {
	Type      string
	Converter ConvertFunction
}

type VerifyFunction func(evidence *Evidence, expectedValidationValues *ValidationValues) error

type Verifier struct {
	Type   string
	Verify VerifyFunction
}

type IssueFunction func(customData []byte) ([]byte, error)

type Issuer struct {
	Type  string
	Issue IssueFunction
}

type Attestation struct {
	Type string `json:"attestation_type"`
	Data string `json:"attestation"`
}

type Evidence struct {
	Type string `json:"attestation_type"`
	Data string `json:"evidence"`
}

type ValidationValues struct {
	Statement []byte
	Mrenclave string
}
