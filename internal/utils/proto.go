/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/discovery"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

func ProtoAsBase64(msg proto.Message) string {
	return base64.StdEncoding.EncodeToString(protoutil.MarshalOrPanic(msg))
}

func ToEndpoint(endpoint string) (*discovery.Endpoint, error) {
	colon := strings.LastIndexByte(endpoint, ':')
	if colon == -1 {
		return nil, fmt.Errorf("invalid format")
	}

	host := endpoint[:colon]
	port, err := strconv.Atoi(endpoint[colon+1:])
	if err != nil {
		return nil, errors.Wrap(err, "invalid port")
	}

	return &discovery.Endpoint{
		Host: host,
		Port: uint32(port),
	}, nil
}

func ExtractEndpoint(credentials *protos.Credentials) (string, error) {
	attestedData := &protos.AttestedData{}
	err := ptypes.UnmarshalAny(credentials.SerializedAttestedData, attestedData)
	if err != nil {
		return "", err
	}

	endpoint := attestedData.HostParams.PeerEndpoint
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port), nil
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
