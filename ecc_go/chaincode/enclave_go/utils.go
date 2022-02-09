/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"crypto/sha256"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
)

func hash(value []byte) []byte {
	h := sha256.New()
	h.Write(value)
	return h.Sum(nil)
}

type fpcIterator struct {
	iterator        shim.StateQueryIteratorInterface
	addReadFunction func(key string, hash []byte)
	decryptFunction func(ciphertext []byte) (plaintext []byte, err error)
}

func newFpcIterator(iterator shim.StateQueryIteratorInterface, addReadFunction func(key string, hash []byte), decryptFunction func(ciphertext []byte) (plaintext []byte, err error)) *fpcIterator {
	return &fpcIterator{
		iterator:        iterator,
		addReadFunction: addReadFunction,
		decryptFunction: decryptFunction,
	}
}

func (i *fpcIterator) HasNext() bool {
	return i.iterator.HasNext()
}

func (i *fpcIterator) Close() error {
	return i.iterator.Close()
}

func (i *fpcIterator) Next() (*queryresult.KV, error) {
	q, err := i.iterator.Next()
	if err != nil {
		return nil, err
	}

	if q == nil {
		return q, nil
	}

	// add to rwset
	i.addReadFunction(utils.TransformToFPCKey(q.Key), hash(q.Value))

	if i.decryptFunction == nil {
		return q, nil
	}

	// decrypt if state decryption function set
	decValue, err := i.decryptFunction(q.Value)
	if err != nil {
		return nil, err
	}

	return &queryresult.KV{
		Namespace: q.Namespace,
		Key:       utils.TransformToFPCKey(q.Key),
		Value:     decValue,
	}, nil
}
