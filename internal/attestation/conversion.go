/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/epid"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/simulation"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("fpc.attestation")

type converterDispatcher struct {
	converters map[string]types.ConvertFunction
}

func newConverterDispatcher() *converterDispatcher {
	return &converterDispatcher{
		converters: make(map[string]types.ConvertFunction),
	}
}

// Register adds new converters to the converterDispatcher
func (d *converterDispatcher) Register(converter ...*types.Converter) error {
	for _, c := range converter {
		if _, ok := d.converters[c.Type]; ok {
			return fmt.Errorf("'%s' type is already registered", c.Type)
		}
		logger.Debugf("Register converter of type '%s'", c.Type)
		d.converters[c.Type] = c.Converter
	}
	return nil
}

// Convert performs the attestation to evidence conversion with help of the registered Converter.
// If there is no matching Converter registered for the input attestation, an error is returned;
// If the invoked ConverterFunction fails, an error is returned; Otherwise an evidence struct is returned.
func (d *converterDispatcher) Convert(attestation *types.Attestation) (*types.Evidence, error) {
	converter, ok := d.converters[attestation.Type]
	if !ok {
		return nil, fmt.Errorf("'%s' type is not registered", attestation.Type)
	}

	logger.Debugf("Invoke converter of type '%s'", attestation.Type)
	evidenceBytes, err := converter([]byte(attestation.Data))
	if err != nil {
		return nil, errors.Wrap(err, "error while converting")
	}

	out := &types.Evidence{
		Type: attestation.Type,
		Data: string(evidenceBytes),
	}

	return out, nil
}

type CredentialConverter struct {
	dispatcher *converterDispatcher
}

func NewDefaultCredentialConverter() *CredentialConverter {
	return NewCredentialConverter(
		simulation.NewSimulationConverter(),
		epid.NewEpidLinkableConverter(),
		epid.NewEpidUnlinkableConverter(),
	)
}

func NewCredentialConverter(converter ...*types.Converter) *CredentialConverter {
	dispatcher := newConverterDispatcher()
	err := dispatcher.Register(converter...)
	if err != nil {
		// ouch this should never happen
		logger.Panicf("cannot create new credential converter! Reason: %s", err.Error())
	}

	return &CredentialConverter{dispatcher: dispatcher}
}

// ConvertCredentials perform attestation evidence conversion (transformation) for a given credentials message (encoded as base64 string)
func (c *CredentialConverter) ConvertCredentials(credentialsOnlyAttestation string) (credentialsWithEvidence string, err error) {
	logger.Debugf("Received Credential: '%s'", credentialsOnlyAttestation)
	credentials, err := utils.UnmarshalCredentials(credentialsOnlyAttestation)
	if err != nil {
		return "", fmt.Errorf("cannot decode credentials: %s", err)
	}

	credentials, err = c.convertCredentials(credentials)
	if err != nil {
		return "", errors.Wrap(err, "cannot convert credentials")
	}

	// marshal "updated" credentials
	credentialsOnlyAttestation = utils.MarshallProtoBase64(credentials)
	logger.Debugf("Converted to Credential: '%s'", credentialsOnlyAttestation)
	return credentialsOnlyAttestation, nil
}

func (c *CredentialConverter) convertCredentials(credentials *protos.Credentials) (*protos.Credentials, error) {
	// get attestation object
	att, err := unmarshalAttestation(credentials.GetAttestation())
	if err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal attestation from credentials")
	}

	// call attestation2evidence conversion
	evidence, err := c.dispatcher.Convert(att)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert attestation")
	}

	evidenceBytes, err := marshalEvidence(evidence)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal evidence")
	}

	// update credentials
	logger.Debugf("evidence: %s\n", evidenceBytes)
	credentials.Evidence = evidenceBytes
	return credentials, nil
}

func unmarshalAttestation(serializedAttestation []byte) (*types.Attestation, error) {
	att := &types.Attestation{}
	err := json.Unmarshal(serializedAttestation, att)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal attestation json")
	}

	return att, nil
}

func marshalEvidence(evidence *types.Evidence) ([]byte, error) {
	evidenceJson, err := json.Marshal(evidence)
	if err != nil {
		return nil, errors.Wrap(err, "json error")
	}
	return evidenceJson, nil
}
