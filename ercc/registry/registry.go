/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

// Package registry implements the client-facing interface of ERCC.
// The corresponding specification can be found in the ERCC Interface section in `$FPC_PATH/docs/design/fabric-v2+/interfaces.md`
package registry

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("ercc")

type Contract struct {
	contractapi.Contract

	Verifier   attestation.Verifier
	IEvaluator utils.IdentityEvaluatorInterface
}

func MyBeforeTransaction(ctx contractapi.TransactionContextInterface) error {
	function, _ := ctx.GetStub().GetFunctionAndParameters()
	logger.Debugf("Invoke [%s]", function)
	return nil
}

// QueryListEnclaveCredentials returns a set of credentials registered for a given chaincode id
// Note: to get the endpoints of FPC endorsing peers do the following:
// - discover all endorsing peers (and their endpoints) for the FPC chaincode using "normal" lifecycle
// - query `getEnclaveId` at all the peers discovered
// - query `queryListEnclaveCredentials` with all received enclave_ids
// this gives you the endpoints and credentials including enclave_vk, and chaincode_ek
//
// Note that this implementation returns a set of (base64-encoded) protobuf-serialized `Credential` objects in order to send it to the receiver.
// That is, the receiver needs to deserialize the return value into []Credentials
// TODO look into custom serializer to make this call more aligned with the FPC API in `interfaces.md`
func (rs *Contract) QueryListEnclaveCredentials(ctx contractapi.TransactionContextInterface, chaincodeId string) ([]string, error) {
	iter, err := ctx.GetStub().GetStateByPartialCompositeKey("namespaces/credentials", []string{chaincodeId})
	if iter != nil {
		defer iter.Close()
	}
	if err != nil {
		return nil, err
	}
	if iter == nil {
		// return empty list, no error
		return nil, nil
	}

	// note that we store serialized credential objects in the chaincode state.
	// since the returns value(s) of this message is sent to the caller (client), it has to serialized anyway.
	// the client needs to deserialize the return value into []Credentials
	var allCredentials []string
	for iter.HasNext() {
		q, err := iter.Next()
		if err != nil {
			return nil, err
		}

		credentialsBase64 := string(q.Value)
		allCredentials = append(allCredentials, credentialsBase64)
	}

	return allCredentials, nil
}

// QueryEnclaveCredentials returns credentials for a provided chaincode and enclave id
func (rs *Contract) QueryEnclaveCredentials(ctx contractapi.TransactionContextInterface, chaincodeId string, enclaveId string) (string, error) {
	key, err := ctx.GetStub().CreateCompositeKey("namespaces/credentials", []string{chaincodeId, enclaveId})
	if err != nil {
		return "", err
	}

	credentialsBase64, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}

	return string(credentialsBase64), nil
}

// QueryListProvisionedEnclaves returns a list of enclave ids of all provisioned enclaves for a given chaincode id. A provisioned enclave is a registered enclave
// that has also the chaincode decryption key.
// Optional Post-MVP;
func (rs *Contract) QueryListProvisionedEnclaves(ctx contractapi.TransactionContextInterface, chaincodeId string) ([]string, error) {

	iter, err := ctx.GetStub().GetStateByPartialCompositeKey("namespaces/credentials", []string{chaincodeId})
	if err != nil {
		return nil, err
	}

	defer func() {
		cerr := iter.Close()
		if err == nil {
			err = cerr
		}
	}()

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

	return enclaveIds, err
}

// QueryChaincodeEndPoints returns the chaincode endpoints for given chaincode id
// (if more than one, they are concatenated with a ",")
func (rs *Contract) QueryChaincodeEndPoints(ctx contractapi.TransactionContextInterface, chaincodeId string) (string, error) {
	iter, err := ctx.GetStub().GetStateByPartialCompositeKey("namespaces/credentials", []string{chaincodeId})
	if iter != nil {
		defer iter.Close()
	}
	if err != nil {
		return "", err
	}
	if iter == nil {
		// return empty list, no error
		return "", nil
	}

	peerEndpoints := ""
	for iter.HasNext() {
		q, err := iter.Next()
		if err != nil {
			return "", err
		}
		credentialsBase64 := string(q.Value)
		credentials, err := utils.UnmarshalCredentials(credentialsBase64)
		if err != nil {
			return "", err
		}

		endpoint, err := utils.ExtractEndpoint(credentials)
		if err != nil {
			return "", err
		}

		if peerEndpoints != "" {
			peerEndpoints = peerEndpoints + "," + endpoint
		} else {
			peerEndpoints = endpoint
		}
	}
	return peerEndpoints, nil
}

// QueryChaincodeEncryptionKey returns the chaincode encryption key for a given chaincode id
func (rs *Contract) QueryChaincodeEncryptionKey(ctx contractapi.TransactionContextInterface, chaincodeId string) (string, error) {
	// NOTE: This is a (momentary) short-cut over the FPC and FPC Lite specification in `docs/design/fabric-v2+/fpc-registration.puml` and `docs/design/fabric-v2+/fpc-key-dist.puml`.  See also `common/enclave/cc_data.cpp` and `protos/fpc/fpc.proto`
	// TODO: remove short cut (see also RegisterEnclave and RegisterCCKeys (Post-MVP)

	// retrieve the enclave id
	iter, err := ctx.GetStub().GetStateByPartialCompositeKey("namespaces/credentials", []string{chaincodeId})
	if iter != nil {
		defer iter.Close()
	}
	if err != nil {
		return "", err
	}

	// pick the first one from the list
	q, err := iter.Next()
	if err != nil {
		return "", err
	}
	_, res, err := ctx.GetStub().SplitCompositeKey(q.Key)
	if err != nil {
		logger.Debugf("no split")
		return "", err
	}
	enclaveId := res[1]

	// recreate composite key of credentials
	k, err := ctx.GetStub().CreateCompositeKey("namespaces/credentials", []string{chaincodeId, enclaveId})
	if err != nil {
		return "", err
	}

	// get credentials from state
	credentialsBase64, err := ctx.GetStub().GetState(k)
	if err != nil {
		return "", err
	}

	// retrieve chaincode ek from credentials
	credentials, err := utils.UnmarshalCredentials(string(credentialsBase64))
	if err != nil {
		return "", err
	}

	var attestedData protos.AttestedData
	if err := credentials.SerializedAttestedData.UnmarshalTo(&attestedData); err != nil {
		return "", err
	}

	chaincodeEKBytes := attestedData.GetChaincodeEk()

	// b64 encoded chaincode key
	b64ChaincodeEK := base64.StdEncoding.EncodeToString(chaincodeEKBytes)
	logger.Debugf("QueryChaincodeEncryptionKey: EK: '%s' / EK b64: '%s'", string(chaincodeEKBytes), b64ChaincodeEK)

	return b64ChaincodeEK, nil
}

// RegisterEnclave register a new FPC chaincode enclave instance
func (rs *Contract) RegisterEnclave(ctx contractapi.TransactionContextInterface, credentialsBase64 string) error {
	logger.Debugf("RegisterEnclave")

	credentials, err := utils.UnmarshalCredentials(credentialsBase64)
	if err != nil {
		return errors.Wrap(err, "invalid credential bytes")
	}

	if len(credentials.Evidence) == 0 {
		return errors.New("evidence is empty")
	}

	// get attested data from credentials
	attestedData, err := utils.UnmarshalAttestedData(credentials.SerializedAttestedData)
	if err != nil {
		return errors.Wrap(err, "invalid attested data message")
	}

	logger.Debugf("- verifying attested data (%s) against evidence (%s)", attestedData.String(), string(credentials.Evidence))
	if err := checkAttestedData(ctx, rs.Verifier, rs.IEvaluator, attestedData, credentials); err != nil {
		return err
	}

	chaincodeId := attestedData.CcParams.ChaincodeId
	enclaveId := utils.GetEnclaveId(attestedData)

	// try create the needed composite key
	key, err := ctx.GetStub().CreateCompositeKey("namespaces/credentials", []string{chaincodeId, enclaveId})
	if err != nil {
		return err
	}

	// check if one enclave is already registered
	registeredCredentialsList, _ := rs.QueryListEnclaveCredentials(ctx, chaincodeId)
	if registeredCredentialsList != nil {
		return fmt.Errorf("an enclave is already registered for chaincode %s", chaincodeId)
	}

	// TODO perform the (enclave) endorsement policy specific tests (MVP/Post-MVP)
	// - MVP (designated chaincode): verify no other enclave (of other peers) is registered or fail
	// - Post-MVP: check consistency with potentially existing enclaves

	// All check passed, now register enclave
	logger.Debugf("Registering credentials at key %s", key)

	if err := ctx.GetStub().PutState(key, []byte(credentialsBase64)); err != nil {
		return fmt.Errorf("cannot store credentials: %s", err)
	}

	// Due to MVP short-cut (see QueryChaincodeEncryptionKey) we already declare chaincode/enclave as provisioned
	// TODO: this has to go to RegisterCCKeys and ImportCCKeys (Post-MVP)
	provisionedKey, err := ctx.GetStub().CreateCompositeKey("namespaces/provisioned", []string{chaincodeId, enclaveId})
	if err != nil {
		return fmt.Errorf("cannot create provisionedKey: %s", err)
	}
	if err := ctx.GetStub().PutState(provisionedKey, []byte("a SignedCCKeyRegistrationMessage")); err != nil {
		return fmt.Errorf("cannot store provisionedKey: %s", err)
	}

	logger.Debugf("RegisterEnclave successful")

	return nil
}

func checkAttestedData(ctx contractapi.TransactionContextInterface, v attestation.Verifier, ie utils.IdentityEvaluatorInterface, attestedData *protos.AttestedData, credentials *protos.Credentials) error {

	// check that the enclave channelId matches ERCC channelId
	if attestedData.CcParams.ChannelId != ctx.GetStub().GetChannelID() {
		return fmt.Errorf("wrong channel! expected=%s, actual=%s", ctx.GetStub().GetChannelID(), attestedData.CcParams.ChannelId)
	}

	// get chaincode definition for chaincode
	ccDef, err := utils.GetChaincodeDefinition(attestedData.CcParams.ChaincodeId, ctx.GetStub())
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
	if err := v.VerifyCredentials(credentials, expectedMrEnclave); err != nil {
		return fmt.Errorf("evidence verification failed: %s", err)
	}

	// next check peer (enclave host) identity is covered by the attestation
	if attestedData.HostParams == nil {
		return errors.New("host params are empty")
	}

	creatorIdentityBytes, err := ctx.GetStub().GetCreator()
	if err != nil {
		return err
	}

	// check that registration transaction creator has same mspid as the enclave owner
	if err := ie.EvaluateCreatorIdentity(creatorIdentityBytes, attestedData.HostParams.PeerMspId); err != nil {
		return fmt.Errorf("creator identity evaluation failed: %s", err)
	}

	// TODO add more checks (POST-MVP)
	// - channel_hash should correspond to peers view of channel id
	// - TLCC_MRENCLAVE matches the version baked into ERCC
	// - validate FPC deployment (restriction) policy

	return nil
}

// RegisterCCKeys  registers a CCKeyRegistration message that confirms that an enclave is provisioned with the chaincode encryption key.
// This method is used during the key generation and key distribution protocol. In particular, during key generation,
// this call sets the chaincode_ek for a chaincode if no chaincode_ek is set yet.
func (rs *Contract) RegisterCCKeys(ctx contractapi.TransactionContextInterface, ccKeyRegistrationMessageBase64 string) error {
	// TODO: Implement me once we remove spec-short-cut,see QueryChaincodeEncryptionKey & RegisterEnclave (Post-MVP)

	//input msg CCKeyRegistrationMessage

	return fmt.Errorf("not implemented yet")
}

// PutKeyExport register key export (Post-MVP feature)
func (rs *Contract) PutKeyExport(ctx contractapi.TransactionContextInterface, exportMessageBase64 string) error {
	// input msg ExportMessage
	// TODO implement me (Post-MVP)
	return fmt.Errorf("not implemented yet")
}

// GetKeyExport retrieve key export (Post-MVP feature)
func (rs *Contract) GetKeyExport(ctx contractapi.TransactionContextInterface, chaincodeId, enclaveId string) (string, error) {
	//input chaincodeId string, enclaveId string
	//return *ExportMessage or  error
	// TODO implement me (Post-MVP)
	return "", fmt.Errorf("not implemented yet")
}
