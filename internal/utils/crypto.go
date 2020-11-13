/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"crypto/sha256"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
)

func ComputedHash(responseMsg *protos.ChaincodeResponseMessage, readset, writeset [][]byte) [32]byte {
	// H(proposal_payload || proposal_signature || response || read set || write set)

	// TODO add missing delimiters or use length encoding;
	// this also needs to be in sync with ecc_enclave/enclave/enclave.cpp
	// https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/ecc_enclave/enclave/enclave.cpp#L85
	h := sha256.New()
	h.Write(responseMsg.Proposal.ProposalBytes)
	h.Write(responseMsg.Proposal.Signature)
	h.Write(responseMsg.EncryptedResponse)
	for _, r := range readset {
		h.Write(r)
	}
	for _, w := range writeset {
		h.Write(w)
	}

	// hash again!!! Note that, sgx_sign() takes the hash, as computed above, as input and hashes again
	return sha256.Sum256(h.Sum(nil))
}
