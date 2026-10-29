/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpcUtils

import (
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// connectPeer creates a gRPC connection to the configured peer
func connectPeer(config *Config) (*grpc.ClientConn, error) {
	var tlsRootCert []byte
	if config.CorePeerTLSEnabled {
		var err error
		tlsRootCert, err = os.ReadFile(filepath.Clean(config.CorePeerTLSRootCertFile))
		if err != nil {
			return nil, errors.Wrap(err, "failed to read TLS root certificate")
		}
	}
	return fpc.NewGrpcConnection(config.CorePeerAddress, tlsRootCert, config.CorePeerId)
}

// peerDialer connects to an enclave peer. Only the configured peer is known, whose enclave is
// registered under its peer id (see `fpcclient init $CORE_PEER_ID`) or its address.
func peerDialer(config *Config) fpc.PeerDialer {
	return func(endpoint string) (grpc.ClientConnInterface, error) {
		if endpoint != config.CorePeerId && endpoint != config.CorePeerAddress {
			return nil, errors.Errorf("enclave peer %s is not the configured peer %s", endpoint, config.CorePeerId)
		}
		return connectPeer(config)
	}
}

// connectGateway connects to the Fabric Gateway of the configured peer using the configured identity
func connectGateway(config *Config) (*client.Gateway, *grpc.ClientConn, identity.Sign, error) {
	conn, err := connectPeer(config)
	if err != nil {
		return nil, nil, nil, err
	}

	id, sign, err := fpc.NewIdentityFromMSP(config.CorePeerLocalMSPID, config.CorePeerMSPConfigPath)
	if err != nil {
		conn.Close()
		return nil, nil, nil, err
	}

	gw, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(conn))
	if err != nil {
		conn.Close()
		return nil, nil, nil, err
	}

	return gw, conn, sign, nil
}
