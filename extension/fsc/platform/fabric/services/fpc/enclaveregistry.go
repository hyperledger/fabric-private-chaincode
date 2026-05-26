/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric"
)

// EnclaveRegistry models the enclave registry
type EnclaveRegistry struct {
	FabricNetworkService *fabric.NetworkService
	Channel              *fabric.Channel
}

// NewEnclaveRegistry returns a new instance of the enclave registry
func NewEnclaveRegistry(fns *fabric.NetworkService, ch *fabric.Channel) *EnclaveRegistry {
	return &EnclaveRegistry{FabricNetworkService: fns, Channel: ch}
}

func (e *EnclaveRegistry) IsAvailable() (bool, error) {
	return e.Channel.Chaincode("ercc").IsAvailable()
}

func (e *EnclaveRegistry) IsPrivate(cid string) (bool, error) {
	_, err := e.ChaincodeEncryptionKey(cid)
	if err != nil && strings.Contains(err.Error(), "no such key") {
		return false, nil
	}
	return err == nil, err
}

// ListProvisionedEnclaves returns the list of provisioned enclaves for the passed chaincode id
func (e *EnclaveRegistry) ListProvisionedEnclaves(cid string) ([]string, error) {
	raw, err := e.Channel.Chaincode("ercc").Query(
		"QueryListProvisionedEnclaves", cid,
	).WithInvokerIdentity(
		e.FabricNetworkService.IdentityProvider().DefaultIdentity(),
	).Call()
	if err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return nil, nil
	}

	var res []string
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, errors.Wrapf(err, "failed unmarshalling [%s:%v]", string(raw), raw)
	}

	return res, nil
}

// ChaincodeEncryptionKey returns the encryption key to be used for the passed chaincode id
func (e *EnclaveRegistry) ChaincodeEncryptionKey(cid string) ([]byte, error) {
	raw, err := e.Channel.Chaincode("ercc").Query(
		"QueryChaincodeEncryptionKey", cid,
	).WithInvokerIdentity(
		e.FabricNetworkService.IdentityProvider().DefaultIdentity(),
	).Call()
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// PeerEndpoints returns the peer endpoints to use when invoking chaincode cid
func (e *EnclaveRegistry) PeerEndpoints(cid string) ([]string, error) {
	raw, err := e.Channel.Chaincode("ercc").Query(
		"QueryChaincodeEndPoints", cid,
	).WithInvokerIdentity(
		e.FabricNetworkService.IdentityProvider().DefaultIdentity(),
	).Call()
	if err != nil {
		return nil, err
	}

	return strings.Split(string(raw), ","), nil
}
