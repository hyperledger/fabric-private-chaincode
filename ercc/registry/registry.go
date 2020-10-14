/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package registry

import (
	"encoding/base64"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
)

type Contract struct {
	contractapi.Contract

	Verifier   attestation.VerifierInterface
	PEvaluator utils.PolicyEvaluatorInterface
}

// returns a set of credentials registered for a given chaincode id
// Note: to get the endpoints of FPC endorsing peers do the following:
// - discover all endorsing peers (and their endpoints) for the FPC chaincode using "normal" lifecycle
// - query `getEnclaveId` at all the peers discovered
// - query `queryListEnclaveCredentials` with all received enclave_ids
// this gives you the endpoints and credentials including enclave_vk, and chaincode_ek
func (rs *Contract) QueryListEnclaveCredentials(ctx contractapi.TransactionContextInterface, chaincodeId string) ([][]byte, error) {
	iter, err := ctx.GetStub().GetStateByPartialCompositeKey("namespaces/credentials", []string{chaincodeId})
	if iter != nil {
		defer iter.Close()
	}
	if err != nil {
		return nil, err
	}

	var allCredentials [][]byte
	for iter.HasNext() {
		q, err := iter.Next()
		if err != nil {
			return nil, err
		}

		allCredentials = append(allCredentials, q.Value)
	}

	return allCredentials, nil
}

func (rs *Contract) QueryEnclaveCredentials(ctx contractapi.TransactionContextInterface, chaincodeId, enclaveId string) ([]byte, error) {
	key, err := ctx.GetStub().CreateCompositeKey("namespaces/credentials", []string{chaincodeId, enclaveId})
	if err != nil {
		return nil, err
	}

	credentials, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}

// Optional Post-MVP;
// returns a list of all provisioned enclaves for a given chaincode id. A provisioned enclave is a registered enclave
// that has also the chaincode decryption key.
func (rs *Contract) QueryListProvisionedEnclaves(ctx contractapi.TransactionContextInterface, chaincodeId string) ([]string, error) {

	iter, err := ctx.GetStub().GetStateByPartialCompositeKey("namespaces/credentials", []string{chaincodeId})
	defer iter.Close()
	if err != nil {
		return nil, err
	}

	var enclaveIds []string

	for iter.HasNext() {
		q, err := iter.Next()
		if err != nil {
			return nil, err
		}

		_, res, err := ctx.GetStub().SplitCompositeKey(q.Key)
		if err != nil {
			return nil, err
		}

		enclaveId := res[1]

		// next check that for each enclaveID there also exists a CCKeyRegistrationMessage
		k, err := ctx.GetStub().CreateCompositeKey("namespaces/provisioned", []string{chaincodeId, enclaveId})
		if err != nil {
			return nil, err
		}

		p, err := ctx.GetStub().GetState(k)
		if err != nil {
			return nil, err
		}

		if p != nil {
			enclaveIds = append(enclaveIds, enclaveId)
		}
	}

	return enclaveIds, nil
}

// returns the chaincode encryption key for a given chaincode id
func (rs *Contract) QueryChaincodeEncryptionKey(ctx contractapi.TransactionContextInterface, chaincodeId string) ([]byte, error) {
	//input chaincodeId string

	//return chaincode_ek []byte
	return nil, nil
}

// register a new FPC chaincode enclave instance
func (rs *Contract) RegisterEnclave(ctx contractapi.TransactionContextInterface, credentialsBase64 string) error {

	credentialBytes, _ := base64.StdEncoding.DecodeString(credentialsBase64)

	var credentials protos.Credentials

	if len(credentialBytes) == 0 {
		return errors.New("credential message is empty")
	}

	if err := proto.Unmarshal(credentialBytes, &credentials); err != nil {
		return errors.Wrap(err, "invalid credential bytes")
	}

	if credentials.SerializedAttestedData == nil {
		return errors.New("attested data is empty")
	}

	if len(credentials.Evidence) == 0 {
		return errors.New("evidence is empty")
	}

	// get attested data from credentials
	var attestedData protos.Attested_Data
	if err := ptypes.UnmarshalAny(credentials.SerializedAttestedData, &attestedData); err != nil {
		return errors.Wrap(err, "invalid attested data message")
	}

	if err := checkAttestedData(ctx, rs.Verifier, rs.PEvaluator, &attestedData, &credentials); err != nil {
		return err
	}

	chaincodeId := attestedData.CcParams.ChaincodeId
	enclaveId := utils.GetEnclaveId(&attestedData)

	key, err := ctx.GetStub().CreateCompositeKey("namespaces/credentials", []string{chaincodeId, enclaveId})
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState(key, credentialBytes); err != nil {
		return fmt.Errorf("cannot store credentials: %s", err)
	}

	return nil
}

func checkAttestedData(ctx contractapi.TransactionContextInterface, v attestation.VerifierInterface, pe utils.PolicyEvaluatorInterface, attestedData *protos.Attested_Data, credentials *protos.Credentials) error {

	// check that the enclave channelId matches ERCC channelId
	if attestedData.CcParams.ChannelId != ctx.GetStub().GetChannelID() {
		return fmt.Errorf("wrong channel! expected=%s, actual=%s", ctx.GetStub().GetChannelID(), attestedData.CcParams.ChannelId)
	}

	// get chaincode definition for chaincode
	ccDef, err := utils.GetChaincodeDefinition(attestedData.CcParams.ChaincodeId, attestedData.CcParams.ChannelId, ctx.GetStub())
	if err != nil {
		return fmt.Errorf("cannot get chaincode definition: %s", err)
	}

	// check that attested data match the chaincode definition
	expectedMrEnclave := ccDef.Version
	if attestedData.CcParams.Version != expectedMrEnclave {
		// note that this is mrenclave
		return fmt.Errorf("mrenclave does not match chaincode definition")
	}

	if attestedData.CcParams.Sequence != ccDef.Sequence {
		return fmt.Errorf("sequence does not match chaincode definition")
	}

	// check that attestation evidence contains expectedMrEnclave as defined in chaincode definition
	if err := v.VerifyEvidence(credentials.Evidence, credentials.SerializedAttestedData.Value, expectedMrEnclave); err != nil {
		return fmt.Errorf("evidence verification failed: %s", err)
	}

	// next check peer (enclave host) identity is covered by the attestation
	if attestedData.HostParams == nil {
		return errors.New("host params are empty")
	}

	if err := pe.EvaluateIdentity(ccDef.ValidationParameter, attestedData.HostParams.PeerIdentity); err != nil {
		return fmt.Errorf("identity does not satisfy endorsement policy: %s", err)
	}

	creatorIdentityBytes, err := ctx.GetStub().GetCreator()
	if err != nil {
		return err
	}

	// check that registration transaction creator has same mspid as the enclave owner
	if err := pe.EvaluateCreatorIdentity(creatorIdentityBytes, attestedData.HostParams.PeerIdentity); err != nil {
		return fmt.Errorf("creator identity evaluation failed: %s", err)
	}

	// TODO add more checks
	// channel_hash should correspond to peers view of channel id (POST-MVP)
	// TLCC_MRENCLAVE matches the version baked into ERCC (POST-MVP)
	// check org-enclave binding (POST-MVP)
	// deployment validation (POST-MVP)
	// validate FPC deployment (restriction) policy

	return nil
}

// registers a CCKeyRegistration message that confirms that an enclave is provisioned with the chaincode encryption key.
// This method is used during the key generation and key distribution protocol. In particular, during key generation,
// this call sets the chaincode_ek for a chaincode if no chaincode_ek is set yet.
func (rs *Contract) RegisterCCKeys(ctx contractapi.TransactionContextInterface, ccKeyRegistrationMessageBase64 string) error {
	//input msg CCKeyRegistrationMessage
	return nil
}

// key distribution (Post-MVP features)
func (rs *Contract) PutKeyExport(ctx contractapi.TransactionContextInterface, exportMessageBase64 string) error {
	// input msg ExportMessage
	return nil
}

func (rs *Contract) GetKeyExport(ctx contractapi.TransactionContextInterface, chaincodeId, enclaveId string) ([]byte, error) {
	//input chaincodeId string, enclaveId string
	//return *ExportMessage or  error
	return nil, nil
}
