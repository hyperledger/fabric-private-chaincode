/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("fpc-client-attest")

type ConvertFunction func(attestationBytes []byte) (evidenceBytes []byte, err error)

type Converter struct {
	Type      string
	Converter ConvertFunction
}

type ConverterDispatcher struct {
	converters map[string]ConvertFunction
}

func NewConverterDispatcher() *ConverterDispatcher {
	return &ConverterDispatcher{
		converters: make(map[string]ConvertFunction),
	}
}

func (d *ConverterDispatcher) Register(converter ...*Converter) error {
	for _, c := range converter {
		if _, ok := d.converters[c.Type]; ok {
			return fmt.Errorf("'%s' type is already registered", c.Type)
		}
		logger.Debugf("Register converter of type '%s'", c.Type)
		d.converters[c.Type] = c.Converter
	}
	return nil
}

func (d *ConverterDispatcher) Convert(attestation *attestation) (*evidence, error) {
	converter, ok := d.converters[attestation.Type]
	if !ok {
		return nil, fmt.Errorf("'%s' type is not registered", attestation.Type)
	}

	logger.Debugf("Invoke converter of type '%s'", attestation.Type)
	evidenceBytes, err := converter([]byte(attestation.Data))
	if err != nil {
		return nil, errors.Wrap(err, "error while converting")
	}

	out := &evidence{
		Type: attestation.Type,
		Data: string(evidenceBytes),
	}

	return out, nil
}

type attestation struct {
	Type string `json:"attestation_type"`
	Data string `json:"attestation"`
}

func unmarshalAttestation(serializedAttestation []byte) (*attestation, error) {
	att := &attestation{}
	err := json.Unmarshal(serializedAttestation, att)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal attestation json")
	}

	return att, nil
}

type evidence struct {
	Type string `json:"attestation_type"`
	Data string `json:"evidence"`
}

type CredentialConverter struct {
	dispatcher *ConverterDispatcher
}

func NewCredentialConverter() *CredentialConverter {
	dispatcher := NewConverterDispatcher()
	err := dispatcher.Register(
		NewSimulationConverter(),
		NewEpidLinkableConverter(),
		NewEpidUnlinkableConverter(),
	)
	if err != nil {
		// ouch this should never happen
		logger.Panicf("cannot create new credential converter! Reason: %s", err.Error())
	}

	return &CredentialConverter{dispatcher: dispatcher}
}

// ConvertCredentials perform attestation evidence conversion (transformation)
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
	credentialsOnlyAttestation = utils.MarshallProto(credentials)
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

	evidenceJson, err := json.Marshal(evidence)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal evidence json")
	}

	// update credentials
	logger.Debugf("evidence: %s\n", evidenceJson)
	credentials.Evidence = evidenceJson
	return credentials, nil
}
