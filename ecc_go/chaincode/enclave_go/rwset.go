/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package enclave_go

import (
	"sync"

	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
)

type ReadWriteSet interface {
	AddRead(key string, hash []byte)
	AddWrite(key string, value []byte)
	AddDelete(key string)
	ToFPCKVSet() *protos.FPCKVSet
}

type read struct {
	kvread *kvrwset.KVRead
	hash   []byte
}

type write struct {
	kvwrite *kvrwset.KVWrite
}

type readWriteSet struct {
	mu     sync.Mutex
	reads  map[string]read
	writes map[string]write
}

func NewReadWriteSet() *readWriteSet {
	return &readWriteSet{
		reads:  make(map[string]read),
		writes: make(map[string]write),
	}
}

func (rwset *readWriteSet) AddRead(key string, hash []byte) {
	rwset.mu.Lock()
	defer rwset.mu.Unlock()
	rwset.reads[key] = read{
		kvread: &kvrwset.KVRead{
			Key:     key,
			Version: nil,
		},
		hash: hash,
	}
}

func (rwset *readWriteSet) AddWrite(key string, value []byte) {
	rwset.mu.Lock()
	defer rwset.mu.Unlock()
	rwset.writes[key] = write{
		kvwrite: &kvrwset.KVWrite{
			Key:      key,
			IsDelete: false,
			Value:    value,
		},
	}
}

func (rwset *readWriteSet) AddDelete(key string) {
	rwset.mu.Lock()
	defer rwset.mu.Unlock()
	rwset.writes[key] = write{
		kvwrite: &kvrwset.KVWrite{
			Key:      key,
			IsDelete: true,
		},
	}
}

func (rwset *readWriteSet) ToFPCKVSet() *protos.FPCKVSet {
	rwset.mu.Lock()
	defer rwset.mu.Unlock()
	fpcKVSet := &protos.FPCKVSet{
		RwSet: &kvrwset.KVRWSet{
			Reads:  []*kvrwset.KVRead{},
			Writes: []*kvrwset.KVWrite{},
		},
		ReadValueHashes: [][]byte{},
	}

	// fill with reads
	for _, read := range rwset.reads {
		fpcKVSet.RwSet.Reads = append(fpcKVSet.RwSet.Reads, read.kvread)
		fpcKVSet.ReadValueHashes = append(fpcKVSet.ReadValueHashes, read.hash)
	}

	// fill with writes
	for _, write := range rwset.writes {
		fpcKVSet.RwSet.Writes = append(fpcKVSet.RwSet.Writes, write.kvwrite)
	}

	return fpcKVSet
}
