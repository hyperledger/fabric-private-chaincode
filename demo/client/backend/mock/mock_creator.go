/*
Copyright Intel Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/msp"
)

func generateMockCreator(mspId string, user string) ([]byte, error) {
	// (1) generate key
	//     NOTE:RSA gen is relative expensive, if turns into problem we
	//     could cache or move to ECDSA
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	// (2) generate (self-signed) x509 certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(100),
		Subject: pkix.Name{
			CommonName:         user,
			OrganizationalUnit: []string{"user", mspId},
			// Below RDNs would be likely in a real certificates but
			// seem to be skipped by dy default in fabric-ca ...
			//   Organization: []string{mspId+".example.com"},
			//   Locality: []string{"San Francisco"},
			//   Province: []string{"California"},
			//   Country: []string{"US"},
		},
	}

	publicKeyCert, err := x509.CreateCertificate(rand.Reader, &template, &template, privateKey.Public(), privateKey)
	if err != nil {
		return nil, err
	}

	// (3) PEM encode certificate
	publicKeyCertPEM := string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: publicKeyCert,
		},
	))

	// (4) encode in protobuf & marshall
	sid := &msp.SerializedIdentity{Mspid: mspId,
		IdBytes: []byte(publicKeyCertPEM)}
	encodedSid, err := proto.Marshal(sid)
	if err != nil {
		return nil, err
	}

	return encodedSid, nil
}
