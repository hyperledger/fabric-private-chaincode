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
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const (
	NonceLength  = 12
	SymKeyLength = 16
	TagLength    = 16
	RSAKeyLength = 3072
)

// GoCrypto implements CSP using pure go
type GoCrypto struct {
}

func NewGoCrypto() *GoCrypto {
	return &GoCrypto{}
}

func (g GoCrypto) NewSymmetricKey() ([]byte, error) {
	keyLength := SymKeyLength
	key := make([]byte, keyLength)
	n, err := rand.Read(key)
	if n != len(key) || err != nil {
		return nil, err
	}

	return key, nil
}

func (g GoCrypto) NewRSAKeys() (publicKey []byte, privateKey []byte, err error) {

	// create rsa private key
	pri, err := rsa.GenerateKey(rand.Reader, RSAKeyLength)
	if err != nil {
		return nil, nil, err
	}

	// serialize
	privateKey = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pri),
	})

	publicKey = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(pri.Public().(*rsa.PublicKey)),
	})

	return publicKey, privateKey, nil
}

func (g GoCrypto) NewECDSAKeys() (publicKey []byte, privateKey []byte, err error) {

	// use secp256r1 (prime256v1)
	curve := elliptic.P256()

	// create ecdsa private key
	pri, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "cannot generate ecdsa key with curve %v", curve)
	}

	// serialize
	x509encodedPri, err := x509.MarshalECPrivateKey(pri)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot serialize private key")
	}

	privateKey = pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: x509encodedPri,
	})

	x509encodedPub, err := x509.MarshalPKIXPublicKey(pri.Public())
	if err != nil {
		return nil, nil, err
	}

	publicKey = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509encodedPub,
	})

	return publicKey, privateKey, nil
}

func (g GoCrypto) VerifyMessage(publicKey []byte, message []byte, signature []byte) error {

	// hash
	hash := sha256.Sum256(message)

	block, _ := pem.Decode(publicKey)
	if block == nil || block.Type != "PUBLIC KEY" {
		return fmt.Errorf("failed to decode PEM block containing public key, got %v", block)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return errors.Wrap(err, "cannot parse public key")
	}

	valid := ecdsa.VerifyASN1(pub.(*ecdsa.PublicKey), hash[:], signature)
	if !valid {
		return fmt.Errorf("failed to verify signature")
	}
	return nil
}

func (g GoCrypto) SignMessage(privateKey []byte, message []byte) (signature []byte, e error) {

	// hash
	hash := sha256.Sum256(message)

	// convert key
	block, _ := pem.Decode(privateKey)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key, got %v", block)
	}

	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// sign

	sig, err := ecdsa.SignASN1(rand.Reader, priv, hash[:])
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (g GoCrypto) PkDecryptMessage(privateKey []byte, encryptedMessage []byte) (message []byte, err error) {
	block, _ := pem.Decode(privateKey)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// using sha1 as used by default with Openssl RSA_PKCS1_OAEP_PADDING
	// https://www.openssl.org/docs/man1.1.1/man3/RSA_public_encrypt.html
	message, err = rsa.DecryptOAEP(sha1.New(), rand.Reader, priv, encryptedMessage, nil)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (g GoCrypto) PkEncryptMessage(publicKey []byte, message []byte) ([]byte, error) {

	block, _ := pem.Decode(publicKey)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// using sha1 as used by default with Openssl RSA_PKCS1_OAEP_PADDING
	// https://www.openssl.org/docs/man1.1.1/man3/RSA_public_encrypt.html
	ciphertext, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, pub, message, nil)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

func (g GoCrypto) DecryptMessage(key []byte, encryptedMessage []byte) ([]byte, error) {

	if len(encryptedMessage) <= NonceLength+TagLength {
		return nil, fmt.Errorf("encrypted message to small. expect len to be larger than %d, actual %d", NonceLength+TagLength, len(encryptedMessage))
	}

	// Note that PDO encryptMessage prepends the nonce and the authentication tag to the ciphertext
	// therefore we extract these values and provide it to aesgcm.Open in the correct format
	nonce := encryptedMessage[:NonceLength]
	tag := encryptedMessage[NonceLength : NonceLength+TagLength]
	ciphertext := encryptedMessage[NonceLength+TagLength:]

	// append tag to ciphertext
	aesgcmCiphertext := ciphertext
	aesgcmCiphertext = append(aesgcmCiphertext, tag...)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, aesgcmCiphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (g GoCrypto) EncryptMessage(key []byte, message []byte) (encryptedMessage []byte, err error) {

	// generate nonce (IV)
	nonce := make([]byte, NonceLength)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	aesgcmCiphertext := aesgcm.Seal(nil, nonce, message, nil)

	// Note that Seal appends the authentication tag to the cipertext, whereas PDO crypto prepends the tag
	ciphertext, tag := aesgcmCiphertext[:len(aesgcmCiphertext)-TagLength], aesgcmCiphertext[len(aesgcmCiphertext)-TagLength:]

	// nonce + tag + cipher
	encryptedMessage = nonce[:]
	encryptedMessage = append(encryptedMessage, tag...)
	encryptedMessage = append(encryptedMessage, ciphertext...)
	return encryptedMessage, nil

}
