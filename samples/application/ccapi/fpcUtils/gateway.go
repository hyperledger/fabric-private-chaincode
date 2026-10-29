/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpcUtils

import (
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// connectGateway connects to the Fabric Gateway of the configured peer using the configured identity
func connectGateway(config *Config) (*client.Gateway, *grpc.ClientConn, error) {
	var tlsRootCert []byte
	if config.CorePeerTLSEnabled {
		var err error
		tlsRootCert, err = os.ReadFile(filepath.Clean(config.CorePeerTLSRootCertFile))
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to read TLS root certificate")
		}
	}

	conn, err := fpc.NewGrpcConnection(config.CorePeerAddress, tlsRootCert, config.CorePeerId)
	if err != nil {
		return nil, nil, err
	}

	id, sign, err := fpc.NewIdentityFromMSP(config.CorePeerLocalMSPID, config.CorePeerMSPConfigPath)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	gw, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(conn))
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return gw, conn, nil
}
