/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// NewGrpcConnection creates a gRPC connection to the Fabric Gateway of a peer.
//
//	Parameters:
//	endpoint is the peer address in format <host>:<port>
//	tlsRootCertPEM is the PEM encoded TLS root certificate of the peer; if empty, TLS is disabled
//	tlsServerName overrides the server name used to verify the peer TLS certificate; if empty, the endpoint host is used
func NewGrpcConnection(endpoint string, tlsRootCertPEM []byte, tlsServerName string) (*grpc.ClientConn, error) {
	transportCredentials := insecure.NewCredentials()
	if len(tlsRootCertPEM) > 0 {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(tlsRootCertPEM) {
			return nil, errors.New("no valid TLS root certificate provided")
		}
		transportCredentials = credentials.NewClientTLSFromCert(certPool, tlsServerName)
	}

	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create gRPC connection to %s", endpoint)
	}
	return conn, nil
}

// NewIdentityFromMSP loads the client identity and its signing function from an MSP folder.
// The folder must contain the PEM encoded certificate in `signcerts` and a single private key in `keystore`.
func NewIdentityFromMSP(mspID, mspPath string) (*identity.X509Identity, identity.Sign, error) {
	certPath, err := findFile(filepath.Join(mspPath, "signcerts"), ".pem")
	if err != nil {
		return nil, nil, err
	}
	certPEM, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read certificate")
	}
	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, nil, err
	}
	id, err := identity.NewX509Identity(mspID, cert)
	if err != nil {
		return nil, nil, err
	}

	keyPath, err := singleFile(filepath.Join(mspPath, "keystore"))
	if err != nil {
		return nil, nil, err
	}
	keyPEM, err := os.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read private key")
	}
	key, err := identity.PrivateKeyFromPEM(keyPEM)
	if err != nil {
		return nil, nil, err
	}
	sign, err := identity.NewPrivateKeySign(key)
	if err != nil {
		return nil, nil, err
	}

	return id, sign, nil
}

// findFile returns the first regular file in dir with the given suffix
func findFile(dir, suffix string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read %s", dir)
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), suffix) {
			return filepath.Join(dir, f.Name()), nil
		}
	}
	return "", errors.Errorf("no file found in %s", dir)
}

// singleFile returns the only file in dir, so that a stray file is never mistaken for the private key.
// Hidden entries and directories are ignored, e.g., the ..data entries of a mounted Kubernetes secret.
func singleFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read %s", dir)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			files = append(files, e.Name())
		}
	}
	if len(files) != 1 {
		return "", errors.Errorf("%s must contain exactly one file, found %d", dir, len(files))
	}
	return filepath.Join(dir, files[0]), nil
}
