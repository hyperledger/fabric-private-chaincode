/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/simulation"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"
)

func Issue(attestedData *anypb.Any) ([]byte, error) {
	issuer := simulation.NewSimulationIssuer()
	att, err := issuer.Issue(attestedData.Value)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get attestation")
	}

	return att, nil
}
