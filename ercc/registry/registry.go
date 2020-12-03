/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

// this package defines the client-facing interface of ERCC as defined in the ERCC Interface section in [specifications](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/docs/design/fabric-v2%2B/interfaces.md)
package registry

import (
	"encoding/base64"
	"fmt"
	"log"

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
	IEvaluator utils.IdentityEvaluatorInterface
}

func MyBeforeTransaction(ctx contractapi.TransactionContextInterface) error {
	function, _ := ctx.GetStub().GetFunctionAndParameters()
	log.Printf("Invoke [%s]", function)
	return nil
}

// returns a set of credentials registered for a given chaincode id
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

	// note that we store serialized credential objects in the chaincode state.
	// since the returns value(s) of this message is sent to the caller (client), it has to serialized anyway.
	// the client needs to deserialize the return value into []Credentials
	var allCredentials []string
	for iter.HasNext() {
		q, err := iter.Next()
		if err != nil {
			return nil, err
		}

		allCredentials = append(allCredentials, base64.StdEncoding.EncodeToString(q.Value))
	}

	return allCredentials, nil
}

func (rs *Contract) QueryEnclaveCredentials(ctx contractapi.TransactionContextInterface, chaincodeId, enclaveId string) (string, error) {
	key, err := ctx.GetStub().CreateCompositeKey("namespaces/credentials", []string{chaincodeId, enclaveId})
	if err != nil {
		return "", err
	}

	credentials, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(credentials), nil
}

// Optional Post-MVP;
// returns a list of enclave ids of all provisioned enclaves for a given chaincode id. A provisioned enclave is a registered enclave
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
func (rs *Contract) QueryChaincodeEncryptionKey(ctx contractapi.TransactionContextInterface, chaincodeId string) (string, error) {
	//  NOTE: This is a (momentary) short-cut over the FPC and FPC Lite specification in `docs/design/fabric-v2+/fpc-registration.puml` and `docs/design/fabric-v2+/fpc-key-dist.puml`.  See also `common/enclave/cc_data.cpp` and `protos/fpc/fpc.proto`

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
		log.Printf("no split")
		return "", err
	}
	enclaveId := res[1]

	// recreate composite key of credentials
	k, err := ctx.GetStub().CreateCompositeKey("namespaces/credentials", []string{chaincodeId, enclaveId})
	if err != nil {
		return "", err
	}

	// get credentials from state
	credentialBytes, err := ctx.GetStub().GetState(k)
	if err != nil {
		return "", err
	}

	// retrieve chaincode ek from credentials
	var credentials protos.Credentials
	if err := proto.Unmarshal(credentialBytes, &credentials); err != nil {
		return "", err
	}

	var attestedData protos.AttestedData
	if err := ptypes.UnmarshalAny(credentials.SerializedAttestedData, &attestedData); err != nil {
		return "", err
	}

	chaincodeEKBytes := attestedData.GetChaincodeEk()

	// b64 encoded chaincode key
	b64ChaincodeEK := base64.StdEncoding.EncodeToString(chaincodeEKBytes)
	log.Printf("QueryChaincodeEncryptionKey:\nEK: %s\nEK b64: %s", string(chaincodeEKBytes), b64ChaincodeEK)

	return b64ChaincodeEK, nil
}

// register a new FPC chaincode enclave instance
func (rs *Contract) RegisterEnclave(ctx contractapi.TransactionContextInterface, credentialsBase64 string) error {
	log.Printf("RegisterEnclave")

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
	var attestedData protos.AttestedData
	if err := ptypes.UnmarshalAny(credentials.SerializedAttestedData, &attestedData); err != nil {
		return errors.Wrap(err, "invalid attested data message")
	}

	log.Printf("- verifying attested data (%s) against evidence (%s)", attestedData.String(), string(credentials.Evidence))
	if err := checkAttestedData(ctx, rs.Verifier, rs.IEvaluator, &attestedData, &credentials); err != nil {
		return err
	}

	chaincodeId := attestedData.CcParams.ChaincodeId
	enclaveId := utils.GetEnclaveId(&attestedData)

	key, err := ctx.GetStub().CreateCompositeKey("namespaces/credentials", []string{chaincodeId, enclaveId})
	if err != nil {
		return err
	}

	// TODO check if this enclave is already registered

	// TODO perform the (enclave) endorsement policy specific tests:
	// - MVP (designated chaincode): verify no other enclave (of other peers) is registered or fail
	// - Post-MVP: check consistency with potentially existing enclaves

	if err := ctx.GetStub().PutState(key, credentialBytes); err != nil {
		return fmt.Errorf("cannot store credentials: %s", err)
	}

	log.Printf("RegisterEnclave successful")

	return nil
}

func checkAttestedData(ctx contractapi.TransactionContextInterface, v attestation.VerifierInterface, ie utils.IdentityEvaluatorInterface, attestedData *protos.AttestedData, credentials *protos.Credentials) error {

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
	if err := v.VerifyEvidence(credentials.Evidence, credentials.SerializedAttestedData.Value, expectedMrEnclave); err != nil {
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

	// TODO add more checks
	// channel_hash should correspond to peers view of channel id (POST-MVP)
	// TLCC_MRENCLAVE matches the version baked into ERCC (POST-MVP)
	// validate FPC deployment (restriction) policy (POST-MVP)

	return nil
}

// registers a CCKeyRegistration message that confirms that an enclave is provisioned with the chaincode encryption key.
// This method is used during the key generation and key distribution protocol. In particular, during key generation,
// this call sets the chaincode_ek for a chaincode if no chaincode_ek is set yet.
func (rs *Contract) RegisterCCKeys(ctx contractapi.TransactionContextInterface, ccKeyRegistrationMessageBase64 string) error {
	// TODO: Implement me for MVP

	//input msg CCKeyRegistrationMessage

	// TODO needs to be implemented for tx args/response encryption

	return fmt.Errorf("not implemented yet")
}

// key distribution (Post-MVP features)
func (rs *Contract) PutKeyExport(ctx contractapi.TransactionContextInterface, exportMessageBase64 string) error {
	// input msg ExportMessage
	// TODO implement me
	return fmt.Errorf("not implemented yet")
}

func (rs *Contract) GetKeyExport(ctx contractapi.TransactionContextInterface, chaincodeId, enclaveId string) (string, error) {
	//input chaincodeId string, enclaveId string
	//return *ExportMessage or  error
	// TODO implement me
	return "", fmt.Errorf("not implemented yet")
}
