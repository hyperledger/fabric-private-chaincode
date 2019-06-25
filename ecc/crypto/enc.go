/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
)

const (
	aesgcm_key_size = 16
	aesgcm_iv_size  = 12
	aesgcm_mac_size = 16
)

func Encrypt(plaintextBytes, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aesgcm_iv_size)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// ciphertext len = plaintext + nonce size (12) + mac size (16)
	// format : nonce | cmac | cipher
	cipherWithMac := aesgcm.Seal(nil, iv, plaintextBytes, nil)

	// extract mac
	cipherLen := len(cipherWithMac) - aesgcm_mac_size
	mac := cipherWithMac[cipherLen:]
	ciphertext := cipherWithMac[:cipherLen]

	out := append(iv, mac...)
	out = append(out, ciphertext...)
	return out, nil
}

func Decrypt(input, key []byte) ([]byte, error) {
	iv := input[:aesgcm_iv_size]
	mac := input[aesgcm_iv_size : aesgcm_iv_size+aesgcm_mac_size]
	ciphertext := input[aesgcm_iv_size+aesgcm_mac_size:]

	ciphertext = append(ciphertext, mac...)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func ParseECDSAPubKey(raw []byte) (*ecdsa.PublicKey, error) {
	pk, err := x509.ParsePKIXPublicKey(raw)
	if err != nil {
		return nil, fmt.Errorf("Failed parsing ecdsa public key [%s]", err)
	}
	enclavePub, ok := pk.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Verification key is not of type ECDSA")
	}
	return enclavePub, nil
}

func GenKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	p256 := elliptic.P256()

	priv, err := ecdsa.GenerateKey(p256, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// transform to sgx pub key format
	pub, ok := priv.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("Ouch!")
	}

	return priv, pub, nil
}

func GenSharedKey(pub *ecdsa.PublicKey, priv *ecdsa.PrivateKey) ([]byte, error) {
	// priv to []byte in big endian
	k := priv.D.Bytes()

	// calcs k*[Bx,By] note that k must be in big endian
	x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, k)

	// fmt.Printf("shared dh\n")
	// fmt.Printf("%s", hex.Dump(x.Bytes()))
	// fmt.Printf("%s", hex.Dump(y.Bytes()))

	// hash the key
	h := sha256.New()
	h.Write(x.Bytes())
	// h.Write(y.Bytes())
	key := h.Sum(nil)

	// fmt.Printf("shared dh hash\n%s", hex.Dump(key))
	return key[:16], nil
}

// this is just for testing ... dont use it
// returns SGX PK format
func GenCrypto(enclavePk []byte) ([]byte, []byte) {
	// transform enclave pk
	enclavePub, err := ParseECDSAPubKey(enclavePk)
	if err != nil {
		panic(fmt.Sprintf("Failed parse pk [%s]", err))
	}

	// gen my keypair
	priv, pub, err := GenKeyPair()
	if err != nil {
		panic(fmt.Sprintf("Failed to generate key pair [%s]", err))
	}

	// transform to sgx pub key format
	pubBytes := make([]byte, 0)
	pubBytes = append(pubBytes, pub.X.Bytes()...)
	pubBytes = append(pubBytes, pub.Y.Bytes()...)

	// gen shared secret
	key, err := GenSharedKey(enclavePub, priv)
	return key, pubBytes
}
