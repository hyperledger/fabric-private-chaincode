/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package epid

import (
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"
	"github.com/pkg/errors"
)

const (
	UnlinkableType = "epid-unlinkable"
	LinkableType   = "epid-linkable"
)

// NewEpidUnlinkableConverter creates a new attestation converter for Intel SGX EPID (unlinkable) attestation
func NewEpidUnlinkableConverter() *types.Converter {
	return &types.Converter{
		Type:      UnlinkableType,
		Converter: newEpidConverter(),
	}
}

// NewEpidLinkableConverter creates a new attestation converter for Intel SGX EPID (linkable) attestation
func NewEpidLinkableConverter() *types.Converter {
	return &types.Converter{
		Type:      LinkableType,
		Converter: newEpidConverter(),
	}
}

func newEpidConverter() types.ConvertFunction {
	return func(attestationBytes []byte) (evidenceBytes []byte, err error) {

		apiKey, err := loadApiKey()
		if err != nil {
			return nil, errors.Wrap(err, "cannot load IAS API key")
		}

		ias := NewIASClient(apiKey)
		evidence, err := ias.RequestAttestationReport(string(attestationBytes))
		if err != nil {
			return nil, errors.Wrap(err, "cannot convert epid attestation")
		}

		return []byte(evidence), nil
	}
}
