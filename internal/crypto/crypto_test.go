package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRSAKeys(t *testing.T) {
	pubKey, privKey, err := NewRSAKeys()
	assert.NoError(t, err)

	assert.NotEmpty(t, pubKey)
	block, _ := pem.Decode(pubKey)
	assert.NotNil(t, block)
	assert.Equal(t, "RSA PUBLIC KEY", block.Type)
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	assert.NotNil(t, pub)
	assert.NoError(t, err)

	assert.NotEmpty(t, privKey)
	block, _ = pem.Decode(privKey)
	assert.NotNil(t, block)
	assert.Equal(t, "RSA PRIVATE KEY", block.Type)
	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	assert.NotNil(t, pri)
	assert.NoError(t, err)
}

func TestNewECDSAKeys(t *testing.T) {
	pubKey, privKey, err := NewECDSAKeys()
	assert.NoError(t, err)

	assert.NotEmpty(t, pubKey)
	block, _ := pem.Decode(pubKey)
	assert.NotNil(t, block)
	assert.Equal(t, "PUBLIC KEY", block.Type)
	// TODO check unsupported curve
	//pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	//assert.NotNil(t, pub)
	//assert.NoError(t, err)

	assert.NotEmpty(t, privKey)
	block, _ = pem.Decode(privKey)
	assert.NotNil(t, block)
	assert.Equal(t, "EC PRIVATE KEY", block.Type)
	//pri, err := x509.ParseECPrivateKey(block.Bytes)
	//assert.NotNil(t, pri)
	//assert.NoError(t, err)
}

func TestNewSymmetricKey(t *testing.T) {
	symKey, err := NewSymmetricKey()
	assert.NotNil(t, symKey)
	assert.NoError(t, err)
}

func TestSignature(t *testing.T) {
	msg := []byte("some message")

	pubKey, privKey, err := NewECDSAKeys()
	assert.NotEmpty(t, pubKey)
	assert.NotEmpty(t, privKey)
	assert.NoError(t, err)

	// fail with invalid key
	sig, err := SignMessage([]byte("invalid key"), msg)
	assert.Nil(t, sig)
	assert.Error(t, err)

	// should succeed
	sig, err = SignMessage(privKey, msg)
	assert.NotNil(t, sig)
	assert.NoError(t, err)

	err = VerifyMessage([]byte("invalid key"), msg, sig)
	assert.Error(t, err)

	err = VerifyMessage(pubKey, []byte("invalid msg"), sig)
	assert.Error(t, err)

	err = VerifyMessage(pubKey, msg, []byte("invalid sig"))
	assert.Error(t, err)

	// should succeed
	err = VerifyMessage(pubKey, msg, sig)
	assert.NoError(t, err)
}

func TestPkEncryption(t *testing.T) {
	msg := []byte("some message")

	pubKey, privKey, err := NewRSAKeys()
	assert.NotEmpty(t, pubKey)
	assert.NotEmpty(t, privKey)
	assert.NoError(t, err)

	cipher, err := PkEncryptMessage([]byte("invalid key"), msg)
	assert.Nil(t, cipher)
	assert.Error(t, err)

	// should succeed
	cipher, err = PkEncryptMessage(pubKey, msg)
	assert.NotNil(t, cipher)
	assert.NoError(t, err)

	plain, err := PkDecryptMessage([]byte("invalid key"), cipher)
	assert.Nil(t, plain)
	assert.Error(t, err)

	// should succeed
	plain, err = PkDecryptMessage(privKey, cipher)
	assert.Equal(t, plain, msg)
	assert.NoError(t, err)
}

func TestSymEncryption(t *testing.T) {
	msg := []byte("some message")

	key, err := NewSymmetricKey()
	assert.NotEmpty(t, key)
	assert.NoError(t, err)

	cipher, err := EncryptMessage([]byte("invalid key"), msg)
	assert.Nil(t, cipher)
	assert.Error(t, err)

	// should succeed
	cipher, err = EncryptMessage(key, msg)
	assert.NotNil(t, cipher)
	assert.NoError(t, err)

	plain, err := DecryptMessage([]byte("invalid key"), cipher)
	assert.Nil(t, plain)
	assert.Error(t, err)

	// should succeed
	plain, err = DecryptMessage(key, cipher)
	assert.Equal(t, plain, msg)
	assert.NoError(t, err)
}
