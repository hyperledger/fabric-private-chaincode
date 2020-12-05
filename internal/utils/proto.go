/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric/protoutil"
)

func MarshallProto(msg proto.Message) string {
	return base64.StdEncoding.EncodeToString(protoutil.MarshalOrPanic(msg))
}

func UnmarshalCredentials(credentialsBase64 string) (*protos.Credentials, error) {
	credentialsBytes, err := base64.StdEncoding.DecodeString(credentialsBase64)
	if err != nil {
		return nil, err
	}

	if len(credentialsBytes) == 0 {
		return nil, fmt.Errorf("credential input empty")
	}

	credentials := &protos.Credentials{}
	err = proto.Unmarshal(credentialsBytes, credentials)
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

// returns enclave_id as hex-encoded string of SHA256 hash over enclave_vk.
func GetEnclaveId(attestedData *protos.AttestedData) string {
	h := sha256.Sum256(attestedData.EnclaveVk)
	return hex.EncodeToString(h[:])
}

func ExtractEndpoint(credentials *protos.Credentials) (string, error) {
	attestedData := &protos.AttestedData{}
	err := ptypes.UnmarshalAny(credentials.SerializedAttestedData, attestedData)
	if err != nil {
		return "", err
	}

	return attestedData.HostParams.PeerEndpoint, nil
}
