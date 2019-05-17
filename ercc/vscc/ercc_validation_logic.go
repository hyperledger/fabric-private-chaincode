/*
* Copyright IBM Corp. All Rights Reserved.
*
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/golang/protobuf/proto"
	commonerrors "github.com/hyperledger/fabric/common/errors"
	"github.com/hyperledger/fabric/common/flogging"
	. "github.com/hyperledger/fabric/core/handlers/validation/api/state"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"

	"github.com/hyperledger-labs/fabric-private-chaincode/ercc"
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	sgxutil "github.com/hyperledger-labs/fabric-private-chaincode/utils"
)

var logger = flogging.MustGetLogger("vscc")

// New creates a new instance of the ercc VSCC
// Typically this will only be invoked once per peer
func New(stateFetcher StateFetcher) *VSCCERCC {
	return &VSCCERCC{
		// ra: &mock.MockVerifier{},
		ra: ercc.GetVerifier(),
		sf: stateFetcher,
	}
}

type VSCCERCC struct {
	ra attestation.Verifier
	sf StateFetcher
}

// Validate validates the given envelope corresponding to a transaction with an endorsement
// policy as given in its serialized form
func (vscc *VSCCERCC) Validate(envelopeBytes []byte, policyBytes []byte) commonerrors.TxValidationError {
	// get the envelope...
	env, err := utils.GetEnvelopeFromBlock(envelopeBytes)
	if err != nil {
		logger.Errorf("ERCC-VSCC error: GetEnvelope failed, err %s", err)
		return policyErr(err)
	}

	// ...and the payload...
	payl, err := utils.GetPayload(env)
	if err != nil {
		logger.Errorf("ERCC-VSCC error: GetPayload failed, err %s", err)
		return policyErr(err)
	}

	// ...and the transaction...
	tx, err := utils.GetTransaction(payl.Data)
	if err != nil {
		logger.Errorf("VSCC error: GetTransaction failed, err %s", err)
		return policyErr(err)
	}

	// loop through each of the actions within
	for _, act := range tx.Actions {
		cap, err := utils.GetChaincodeActionPayload(act.Payload)
		if err != nil {
			logger.Errorf("VSCC error: GetChaincodeActionPayload failed, err %s", err)
			return policyErr(err)
		}

		pRespPayload, err := utils.GetProposalResponsePayload(cap.Action.ProposalResponsePayload)
		if err != nil {
			logger.Errorf("VSCC error: GetProposalResponsePayload failed, err %s", err)
			return policyErr(err)
		}

		ccAction := &peer.ChaincodeAction{}
		err = proto.Unmarshal(pRespPayload.Extension, ccAction)
		if err != nil {
			logger.Errorf("VSCC error: GetProposalResponsePayload failed, err %s", err)
			return policyErr(err)
		}

		err = vscc.checkAttestation(ccAction)
		if err != nil {
			logger.Errorf("VSCC error: checkAttestation failed, err %s", err)
			return policyErr(err)
		}
	}
	return nil
}

func (t *VSCCERCC) checkAttestation(respPayload *peer.ChaincodeAction) error {
	logger.Debug("checkEnclaveEndorsement starts")

	var err error

	txRWSet := &rwsetutil.TxRwSet{}
	if err = txRWSet.FromProtoBytes(respPayload.Results); err != nil {
		return err
	}

	for _, ns := range txRWSet.NsRwSets {
		logger.Debugf("Namespace %s", ns.NameSpace)

		// TODO make this more flexible
		if ns.NameSpace != "ercc" {
			continue
		}

		writes := ns.KvRwSet.Writes
		if len(writes) != 1 {
			return errors.New("Expected one write")
		}
		write := writes[0]

		logger.Debugf("checkEnclaveEndorsement info: validating key %s", write.Key)

		attestationReport := attestation.IASAttestationReport{}
		err = json.Unmarshal(write.Value, &attestationReport)
		if err != nil {
			return fmt.Errorf("txRWSet.Unmarshal failed, err %s", err)
		}

		// transform INTEL pk to DER format
		block, _ := pem.Decode([]byte(attestation.IntelPubPEM))

		// transform sig-pk from attestation report to DER format
		verificationPK, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("x509.ParsePKIXPublicKey failed, err: %s", err)
		}

		// verify attestation report
		isValid, err := t.ra.VerifyAttestationReport(verificationPK, attestationReport)
		if err != nil {
			return fmt.Errorf("VerifyAttestationReport failed, err %s", err)
		}
		if !isValid {
			return errors.New("Attestation report is not valid")
		}
		logger.Debugf("Attestation valid!")

		// verify write.Key
		enclavePkHash := sha256.Sum256(attestationReport.EnclavePk)
		if write.Key != base64.StdEncoding.EncodeToString(enclavePkHash[:]) {
			return errors.New("Error: write.Key does not match enclave public key hash from attestation")
		}
		logger.Debugf("write.Key correct!")

		// verify that pk attestation report matches the one in the quote
		isValid, err = t.ra.CheckEnclavePkHash(attestationReport.EnclavePk, attestationReport)
		if err != nil {
			return fmt.Errorf("Error while checking enclave PK: %s", err)
		}
		if !isValid {
			return errors.New(" Enclave PK does not match attestation report!")
		}
		logger.Debugf("Enclave PK matches attestation report!")

		channelState, err := t.sf.FetchState()
		if err != nil {
			return fmt.Errorf("Fetch channel state failed, err %s", err)
		}
		defer channelState.Done()

		state := &state{channelState}
		// get mrenclave from ledger
		// FIXME: remove hardcoding of those strings
		mrenclave, err := state.GetState("ecc", sgxutil.MrEnclaveStateKey)
		if err != nil {
			return errors.New("mrenclave does not exist")
		}
		if mrenclave == nil {
			return errors.New("mrenclave is empty")
		}
		logger.Debugf("mrenclave from ecc: %s", mrenclave)

		// check mrenclave
		matches, err := t.ra.CheckMrEnclave(string(mrenclave), attestationReport)
		if err != nil {
			return fmt.Errorf("Error while attestation report verification: %s", err)
		}
		if !matches {
			logger.Errorf("Expected MRENCLAVE: %s", string(mrenclave))
			return errors.New("Attestation report does not match MRENCLAVE!")
		}
		logger.Debugf("mrenclave matches attestation report!")
	}

	return nil
}

type state struct {
	State
}

// GetState retrieves the value for the given key in the given namespace
func (s *state) GetState(namespace string, key string) ([]byte, error) {
	values, err := s.GetStateMultipleKeys(namespace, []string{key})
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	return values[0], nil
}

func policyErr(err error) *commonerrors.VSCCEndorsementPolicyError {
	return &commonerrors.VSCCEndorsementPolicyError{
		Err: err,
	}
}
