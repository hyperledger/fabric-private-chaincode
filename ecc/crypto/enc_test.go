/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestDH(t *testing.T) {
	p256 := elliptic.P256()

	// create enclave priv, pk just for testing
	enclavePriv, err := ecdsa.GenerateKey(p256, rand.Reader)
	enclavePub, ok := enclavePriv.Public().(*ecdsa.PublicKey)
	if !ok {
		t.Error("cannot cast ecdsa pub key", err)
	}

	priv, pub, err := GenKeyPair()
	if err != nil {
		t.Error("cannot generate key pair", err)
	}
	pubBytes := make([]byte, 0)
	pubBytes = append(pubBytes, pub.X.Bytes()...)
	pubBytes = append(pubBytes, pub.Y.Bytes()...)

	// gen shared secret
	key, _ := GenSharedKey(enclavePub, priv)

	plaintext := []byte("Moin")
	ciphertext, _ := Encrypt(plaintext, key)
	fmt.Printf("Base64 cipher: %s\n", ciphertext)
	fmt.Printf("Base64 my pk: %s\n", base64.StdEncoding.EncodeToString(pubBytes))
}
