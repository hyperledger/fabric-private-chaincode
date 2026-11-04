//go:build fpc_cgo

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"github.com/hyperledger/fabric-private-chaincode/sgx/attestation/epid/pdo"
)

func init() {
	registry.add(pdo.NewEpidLinkableVerifier())
	registry.add(pdo.NewEpidUnlinkableVerifier())
}
