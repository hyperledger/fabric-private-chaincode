//go:build WITH_PDO_CRYPTO

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/epid/pdo"
)

func init() {
	registry.add(pdo.NewEpidLinkableVerifier())
	registry.add(pdo.NewEpidUnlinkableVerifier())
}
