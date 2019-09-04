/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"

	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("ecdsa")

type ecdsaSignature struct {
	R *big.Int
	S *big.Int
}

func UnmarshalECDSASignature(raw []byte) (*big.Int, *big.Int, error) {
	// Unmarshal
	sig := new(ecdsaSignature)
	if _, err := asn1.Unmarshal(raw, sig); err != nil {
		return nil, nil, fmt.Errorf("Failed unmashalling signature [%s]", err)
	}

	// Validate sig
	if sig.R == nil {
		return nil, nil, errors.New("Invalid signature. R must be different from nil")
	}
	if sig.S == nil {
		return nil, nil, errors.New("Invalid signature. S must be different from nil")
	}

	if sig.R.Sign() != 1 {
		return nil, nil, errors.New("Invalid signature. R must be larger than zero")
	}
	if sig.S.Sign() != 1 {
		return nil, nil, errors.New("Invalid signature. S must be larger than zero")
	}

	return sig.R, sig.S, nil
}

func MarshalEnclaveSignature(input []byte) ([]byte, error) {

	r := new(big.Int)
	r.SetBytes(input[:32])

	s := new(big.Int)
	s.SetBytes(input[32:])

	// Validate sig
	if r == nil {
		return nil, errors.New("Invalid signature. R must be different from nil")
	}
	if r == nil {
		return nil, errors.New("Invalid signature. S must be different from nil")
	}

	if r.Sign() != 1 {
		return nil, errors.New("Invalid signature. R must be larger than zero")
	}
	if s.Sign() != 1 {
		return nil, errors.New("Invalid signature. S must be larger than zero")
	}

	return asn1.Marshal(ecdsaSignature{r, s})
}

// UnmarshallEnclavePk converts DER-encoded PKIX format to sgx format big endian
func UnmarshalEnclavePk(input []byte) ([]byte, error) {

	re, err := x509.ParsePKIXPublicKey(input)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse DER encoded public key [%s]", err)
	}

	pub := re.(*ecdsa.PublicKey)

	out := make([]byte, 64)
	copy(out[:32], pub.X.Bytes())
	copy(out[32:], pub.Y.Bytes())

	return out, nil
}

func EnclavePk2ECDSAPK(input []byte) (*ecdsa.PublicKey, error) {
	x := new(big.Int)
	x.SetBytes(input[:32])

	y := new(big.Int)
	y.SetBytes(input[32:])

	// check that PK is a valid for NIST P-256 elliptic curve
	curve := elliptic.P256()
	if !curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("Public key not valid (Point not on curve)")
	}

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

// MarshallEnclavePk converts sgx format Big endian to DER-encoded PKIX format
func MarshalEnclavePk(input []byte) (raw []byte, err error) {
	pk, err := EnclavePk2ECDSAPK(input)
	if err != nil {
		return nil, err
	}

	// serialize pk to DER-encoded PKIX format
	raw, err = x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		return nil, fmt.Errorf("Failed marshalling key [%s]", err)
	}

	return raw, nil
}

// ECDSAVerifier implements Verifier interface!
type ECDSAVerifier struct {
}

// Verify returns true if signature validation of enclave return is correct; other false
func (v *ECDSAVerifier) Verify(txType, encoded_args, responseData []byte, readset, writeset [][]byte, signature, enclavePk []byte) (bool, error) {
	// unmarshall signature
	r, s, err := UnmarshalECDSASignature(signature)
	if err != nil {
		return false, fmt.Errorf("Failed unmarshalling signature [%s]", err)
	}

	// unmarshall pk
	pk, err := x509.ParsePKIXPublicKey(enclavePk)
	if err != nil {
		return false, fmt.Errorf("Failed parsing ecdsa public key [%s]", err)
	}
	ecdsaPublickey, ok := pk.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("Verification key is not of type ECDSA")
	}

	// Note: below signature was created in ecc_enclave/enclave/enclave.cpp::gen_response
	// see also replicated verification in tlcc_enclave/enclave/ledger.cpp::int parse_endorser_transaction (for TLCC)

	// H(txType in {"init", "invoke"} || encoded_args || response || read set || write set)
	h := sha256.New()
	logger.Debugf("txType: %s\n", string(txType))
	h.Write(txType)
	logger.Debugf("encoded_args: %s\n", string(encoded_args))
	h.Write(encoded_args)
	logger.Debugf("response data len: %d\n", len(responseData))
	h.Write(responseData)
	logger.Debugf("write_set: ")
	for _, r := range readset {
		logger.Debugf("\\-> %s\n", string(r))
		h.Write(r)
	}
	logger.Debugf("write_set: ")
	for _, w := range writeset {
		logger.Debugf("\\-> %s\n", string(w))
		h.Write(w)
	}
	hash := h.Sum(nil)

	hashBase64 := base64.StdEncoding.EncodeToString(hash)
	logger.Debugf("ecc sig hash (base64): %s\n", hashBase64)

	// hash again!!! Note that, sgx_sign() takes the hash, as computed above, as input and hashes again
	hash2 := sha256.Sum256(hash)

	return ecdsa.Verify(ecdsaPublickey, hash2[:], r, s), nil
}
