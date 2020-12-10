/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package enclave

import (
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("enclave")

type StubInterface interface {

	// triggered by an admin
	Init(chaincodeParams, hostParams, attestationParams []byte) ([]byte, error)

	// key generation
	GenerateCCKeys() (*protos.SignedCCKeyRegistrationMessage, error)

	// key distribution (Post-MVP Feature)
	ExportCCKeys(credentials *protos.Credentials) (*protos.SignedExportMessage, error)
	ImportCCKeys() (*protos.SignedCCKeyRegistrationMessage, error)

	// returns the EnclaveId hosted by the peer
	GetEnclaveId() (string, error)

	// chaincode invoke
	ChaincodeInvoke(stub shim.ChaincodeStubInterface) ([]byte, error)
}
