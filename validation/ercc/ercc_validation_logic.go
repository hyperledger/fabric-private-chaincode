/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// This ercc-vscc code is deprecated and will be integrated in ercc with the refactoring

package ercc

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/peer"
	commonerrors "github.com/hyperledger/fabric/common/errors"
	"github.com/hyperledger/fabric/common/flogging"
	validation "github.com/hyperledger/fabric/core/handlers/validation/api/state"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("ercc-vscc")

// New creates a new instance of the ercc VSCC
// Typically this will only be invoked once per peer
func New(stateFetcher validation.StateFetcher) *VSCCERCC {
	return &VSCCERCC{
		ra: attestation.GetVerifier(),
		sf: stateFetcher,
	}
}

type VSCCERCC struct {
	ra attestation.Verifier
	sf validation.StateFetcher
}

// Validate validates the given envelope corresponding to a transaction with an endorsement
// policy as given in its serialized form
func (vscc *VSCCERCC) Validate(envelopeBytes []byte, policyBytes []byte) commonerrors.TxValidationError {
	// get the envelope...
	env, err := protoutil.GetEnvelopeFromBlock(envelopeBytes)
	if err != nil {
		logger.Errorf("ERCC-VSCC error: GetEnvelope failed, err %s", err)
		return policyErr(err)
	}

	// ...and the payload...
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		logger.Errorf("ERCC-VSCC error: GetPayload failed, err %s", err)
		return policyErr(err)
	}

	// ...and the transaction...
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		logger.Errorf("VSCC error: GetTransaction failed, err %s", err)
		return policyErr(err)
	}

	// loop through each of the actions within
	for _, act := range tx.Actions {
		cap, err := protoutil.UnmarshalChaincodeActionPayload(act.Payload)
		if err != nil {
			logger.Errorf("VSCC error: GetChaincodeActionPayload failed, err %s", err)
			return policyErr(err)
		}

		pRespPayload, err := protoutil.UnmarshalProposalResponsePayload(cap.Action.ProposalResponsePayload)
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

		// next get invocation specs
		cpp, err := protoutil.UnmarshalChaincodeProposalPayload(cap.ChaincodeProposalPayload)
		if err != nil {
			return policyErr(err)
		}

		cis := &peer.ChaincodeInvocationSpec{}
		if err = proto.Unmarshal(cpp.Input, cis); err != nil {
			return policyErr(err)
		}

		function := cis.ChaincodeSpec.Input.Args[0]
		logger.Debugf("function: %s\n", string(function))

		// we only perform attestation checks for registerEnclave invocations triggered through __setup of ecc
		if string(function) == "__setup" {
			err = vscc.checkAttestation(ccAction)
			if err != nil {
				logger.Errorf("VSCC error: checkAttestation failed, err %s", err)
				return policyErr(err)
			}
		}
	}
	return nil
}

func (t *VSCCERCC) checkAttestation(respPayload *peer.ChaincodeAction) error {
	logger.Debug("checkAttestation starts")

	var err error

	txRWSet := &rwsetutil.TxRwSet{}
	if err = txRWSet.FromProtoBytes(respPayload.Results); err != nil {
		return err
	}

	// ensure that mrenclave is read from ecc namespace
	eccNamespace, err := getEccNamespace(txRWSet)
	if err != nil {
		return err
	}
	logger.Debugf("eccNamespace found: %s", eccNamespace)

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

		if block == nil {
			return fmt.Errorf("public key resides in a malformed PEM block")
		}

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
		mrenclave, err := state.GetState(eccNamespace, utils.MrEnclaveStateKey)
		if err != nil {
			return fmt.Errorf("getting MRENCLAVE failed, err %s", err)
		}
		// if mrenclave not already on the ledger
		if mrenclave == nil {
			// it must be in the writeset
			mrenclave = getMrEnclaveWriteSet(eccNamespace, txRWSet)
			if mrenclave == nil {
				return fmt.Errorf("no MRENCLAVE on the ledger nor in write set for %s", eccNamespace)
			}
		}
		logger.Debugf("mrenclave for %s: %s", eccNamespace, mrenclave)

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
	validation.State
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

func getEccNamespace(txRWSet *rwsetutil.TxRwSet) (string, error) {
	// Readsets must contain only one *some_fpc_chaincode_name*.MRENCLAVE
	eccNamespace := ""
	for _, ns := range txRWSet.NsRwSets {
		for _, read := range ns.KvRwSet.Reads {
			if read.Key == utils.MrEnclaveStateKey {
				logger.Debugf("found MRENCLAVE within namespace: %s", ns.NameSpace)
				if eccNamespace != "" {
					return "", fmt.Errorf("mutiple namespaces with MRENCLAVE key found")
				}
				eccNamespace = ns.NameSpace
			}
		}
	}
	if eccNamespace == "" {
		return "", fmt.Errorf("no ecc namespace found")
	}
	return eccNamespace, nil
}

func getMrEnclaveWriteSet(targetNameSpace string, txRWSet *rwsetutil.TxRwSet) []byte {
	for _, ns := range txRWSet.NsRwSets {
		if ns.NameSpace == targetNameSpace {
			for _, write := range ns.KvRwSet.Writes {
				if write.Key == utils.MrEnclaveStateKey {
					return write.Value
				}
			}
		}
	}
	return nil
}
