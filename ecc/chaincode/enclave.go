/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type Enclave interface {

	// Init initializes the chaincode enclave.
	// The input and output parameters are serialized protobufs
	// triggered by an admin
	Init(chaincodeParams, hostParams, attestationParams []byte) (credentials []byte, err error)

	// GetEnclaveId returns the EnclaveId hosted by the peer
	GetEnclaveId() (string, error)

	// key distribution (Post-MVP Feature)

	// GenerateCCKeys returns a signed CCKeyRegistration Message including
	// The output parameters is a serialized protobuf
	GenerateCCKeys() (signedCCKeyRegistrationMessage []byte, err error)

	// ExportCCKeys exports chaincode secrets to enclave with provided credentials
	// The input and output parameters are serialized protobufs
	ExportCCKeys(credentials []byte) (signedExportMessage []byte, err error)

	// ImportCCKeys imports chaincode secrets
	// The output parameters is a serialized protobuf
	ImportCCKeys() (signedCCKeyRegistrationMessage []byte, err error)

	// ChaincodeInvoke invokes fpc chaincode inside enclave
	// chaincodeRequestMessage and chaincodeResponseMessage are serialized protobuf
	ChaincodeInvoke(stub shim.ChaincodeStubInterface, chaincodeRequestMessage []byte) (chaincodeResponseMessage []byte, err error)
}
