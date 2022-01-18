/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/hyperledger/fabric-private-chaincode/internal/crypto"
)

type EnclaveIdentity struct {
	csp        crypto.CSP
	privateKey []byte
	publicKey  []byte
	enclaveId  string
}

type EnclaveIdentityFunctions interface {
	Sign(msg []byte) (signature []byte, err error)
	GetPublicKey() []byte
	GetEnclaveId() (string, error)
}

func NewEnclaveIdentity(csp crypto.CSP) (*EnclaveIdentity, error) {
	var err error
	e := &EnclaveIdentity{}
	e.csp = csp

	// create enclave keys
	e.publicKey, e.privateKey, err = csp.NewECDSAKeys()
	if err != nil {
		return nil, err
	}

	// calculate enclave id
	pubHash := sha256.Sum256(e.publicKey)
	e.enclaveId = strings.ToUpper(hex.EncodeToString(pubHash[:]))

	return e, nil
}

func (e *EnclaveIdentity) Sign(msg []byte) (signature []byte, err error) {
	signature, err = e.csp.SignMessage(e.privateKey, msg)
	return
}

func (e *EnclaveIdentity) GetPublicKey() []byte {
	return e.publicKey
}

func (e *EnclaveIdentity) GetEnclaveId() string {
	return e.enclaveId
}

type ChaincodeIdentity struct {
	csp          crypto.CSP
	ccPrivateKey []byte
	ccPublicKey  []byte
	stateKey     []byte
}

type ChaincodeIdentityFunctions interface {
	GetPublicKey() []byte
	PkDecryptMessage(ciphertext []byte) (plaintext []byte, err error)
	StateEncryptionFunctions
}

type StateEncryptionFunctions interface {
	EncryptState(plaintext []byte) (ciphertext []byte, err error)
	DecryptState(ciphertext []byte) (plaintext []byte, err error)
}

func NewChaincodeIdentity(csp crypto.CSP) (*ChaincodeIdentity, error) {
	var err error
	c := &ChaincodeIdentity{}
	c.csp = csp

	// create chaincode encryption keys
	c.ccPublicKey, c.ccPrivateKey, err = csp.NewRSAKeys()
	if err != nil {
		return nil, err
	}

	// create state key
	c.stateKey, err = csp.NewSymmetricKey()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *ChaincodeIdentity) GetPublicKey() []byte {
	return c.ccPublicKey
}

func (c *ChaincodeIdentity) PkDecryptMessage(ciphertext []byte) (plaintext []byte, err error) {
	return c.csp.PkDecryptMessage(c.ccPrivateKey, ciphertext)
}

func (c *ChaincodeIdentity) EncryptState(plaintext []byte) (ciphertext []byte, err error) {
	return c.csp.EncryptMessage(c.stateKey, plaintext)
}

func (c *ChaincodeIdentity) DecryptState(ciphertext []byte) (plaintext []byte, err error) {
	return c.csp.DecryptMessage(c.stateKey, ciphertext)

}
