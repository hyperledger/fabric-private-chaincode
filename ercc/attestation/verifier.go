/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"reflect"
)

// IASRequestBody sent to IAS (Intel attestation service)
type IASRequestBody struct {
	Quote string `json:"isvEnclaveQuote"`
}

// EnclaveQuote is a struct for a quote object. This object is produced by SGX
type EnclaveQuote struct {
	Version     uint16
	SignType    uint16
	EPIDGroupID [4]byte
	QeSVN       [2]byte
	PceSVN      [2]byte
	XeID        uint32
	Basename    [32]byte
	// ReportBody  ReportBodyT
	CPUSVN     [16]byte
	MiscSelect [4]byte
	Reserved1  [28]byte
	Attributes [16]byte
	MrEnclave  [32]byte
	Reserved2  [32]byte
	MrSigner   [32]byte
	Reserved3  [96]byte
	ISVProdID  [2]byte
	ISVSVN     [2]byte
	Reserved4  [60]byte
	ReportData [64]byte
	//SignatureLen uint32
	//Signature    []byte
}

// QuoteFromBytes parses a byte string to EnclaveQuote
func QuoteFromBytes(quoteAsBytes []byte) (EnclaveQuote, error) {
	quote := EnclaveQuote{}
	err := binary.Read(bytes.NewReader(quoteAsBytes), binary.LittleEndian, &quote)
	if err != nil {
		return quote, err
	}
	return quote, nil
}

// QuoteFromBase64 parses a byte string to EnclaveQuote
func QuoteFromBase64(quoteBase64 string) (EnclaveQuote, error) {
	quoteAsBytes, err := base64.StdEncoding.DecodeString(quoteBase64)
	if err != nil {
		return EnclaveQuote{}, err
	}
	return QuoteFromBytes(quoteAsBytes)
}

func QuoteFromAttestationReport(report IASAttestationReport) (EnclaveQuote, error) {
	reportBody := IASReportBody{}
	err := json.Unmarshal(report.IASReportBody, &reportBody)
	if err != nil {
		return EnclaveQuote{}, err
	}

	quote, err := QuoteFromBase64(reportBody.IsvEnclaveQuoteBody)
	if err != nil {
		return EnclaveQuote{}, err
	}
	return quote, nil
}

// ReportBodyT contains report body
// type ReportBodyT struct {
// 	CPUSVN     [16]byte
// 	MiscSelect [4]byte
// 	Reserved1  [28]byte
// 	Attributes [16]byte
// 	MrEnclave  [32]byte
// 	Reserved2  [32]byte
// 	MrSigner   [32]byte
// 	Reserved3  [96]byte
// 	ISVProdID  [2]byte
// 	ISVSVN     [2]byte
// 	Reserved4  [60]byte
// 	ReportData [64]byte
// }

// Verifier interface
type Verifier interface {
	VerifyAttestationReport(verificationPubKey interface{}, report IASAttestationReport) (bool, error)
	CheckMrEnclave(mrEnclaveHexString string, report IASAttestationReport) (bool, error)
	CheckEnclavePkHash(pkBytes []byte, report IASAttestationReport) (bool, error)
}

// VerifierImpl implements Verifier interface!
type VerifierImpl struct {
}

// VerifyAttestationReport verifies IASAttestationReport signature using provided verification key
func (v *VerifierImpl) VerifyAttestationReport(verificationPubKey interface{}, report IASAttestationReport) (bool, error) {

	// decode certs
	certs, _ := url.QueryUnescape(report.IASReportSigningCertificate)

	// read signing cert first
	block, rest := pem.Decode([]byte(certs))
	if block == nil {
		return false, errors.New("provided cert not PEM formatted")
	}

	signCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, errors.New("failed to parse signing certificate:" + err.Error())
	}

	// read ca cert
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(rest); !ok {
		return false, errors.New("Failed to parse root certificate")
	}

	opts := x509.VerifyOptions{
		Roots: roots,
	}

	// verify signing Cert
	if _, err := signCert.Verify(opts); err != nil {
		return false, errors.New("Failed to verify signing certificate")
	}

	// verify response signature
	signature, _ := base64.StdEncoding.DecodeString(report.IASReportSignature)
	hashedBody := sha256.Sum256(report.IASReportBody)

	// check verification if its rsa key
	rsaPublickey, ok := verificationPubKey.(*rsa.PublicKey)
	if !ok {
		return false, errors.New("Verification key is not of type RSA")
	}

	// if err = rsa.VerifyPKCS1v15(signCertPK, crypto.SHA256, hashedBody[:], signature); err != nil {
	if err = rsa.VerifyPKCS1v15(rsaPublickey, crypto.SHA256, hashedBody[:], signature); err != nil {
		return false, errors.New("Signature verification failed: " + err.Error())
	}

	return true, nil
}

// CheckMrEnclave returns true if mrenclave in attestation report matches the expected value. Expected value input as base64.
func (v *VerifierImpl) CheckMrEnclave(mrEnclaveHexString string, report IASAttestationReport) (bool, error) {

	quote, err := QuoteFromAttestationReport(report)
	if err != nil {
		return false, err
	}

	mrenclave, err := hex.DecodeString(mrEnclaveHexString)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(mrenclave[:32], quote.MrEnclave[:32]), nil
}

// CheckEnclavePkHash returns true if hash of enclave pk in quote matches the expected value.
func (v *VerifierImpl) CheckEnclavePkHash(pkBytes []byte, report IASAttestationReport) (bool, error) {

	quote, err := QuoteFromAttestationReport(report)
	if err != nil {
		return false, err
	}

	pub, err := x509.ParsePKIXPublicKey(pkBytes)
	if err != nil {
		return false, fmt.Errorf("x509.ParsePKIXPublicKey error %s", err)
	}

	ecdsaPublickey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("enclave key is not ecdsa key")
	}

	h := sha256.New()
	h.Write(ecdsaPublickey.X.Bytes())
	h.Write(ecdsaPublickey.Y.Bytes())
	enclavePkHash := h.Sum(nil)

	return reflect.DeepEqual(enclavePkHash[:32], quote.ReportData[:32]), nil
}
