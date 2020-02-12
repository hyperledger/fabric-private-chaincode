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
	"fmt"
	"math/big"
	"regexp"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/msp"
)

func generateMockCreatorPKIXName(mspId string, org string, user string) pkix.Name {
	return pkix.Name{
		CommonName:         user,
		OrganizationalUnit: []string{"user", org},
		// Below RDNs would be likely in a real certificates but
		// seem to be skipped by dy default in fabric-ca ...
		//   Organization: []string{mspId+".example.com"},
		//   Locality: []string{"San Francisco"},
		//   Province: []string{"California"},
		//   Country: []string{"US"},
	}
}

func generateMockCreatorDN(mspId string, org string, user string) string {
	return generateMockCreatorPKIXName(mspId, org, user).String()
}

func parseCreatorDN(dn string) (org string, user string, err error) {
	pat := `CN=([^,]+),OU=([^,+]+)\+OU=([^,+]+)`
	re := regexp.MustCompile(pat)

	if re.MatchString(dn) {
		matches := re.FindStringSubmatch(dn)
		user = matches[1]
		// role = matches[2]
		org = matches[3]
	} else {
		err = fmt.Errorf("dn '%s' did not match pattern '%v'", dn, pat)
	}
	return
}

func generateMockCreator(mspId string, org string, user string) ([]byte, error) {
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
		Subject:      generateMockCreatorPKIXName(mspId, org, user),
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
