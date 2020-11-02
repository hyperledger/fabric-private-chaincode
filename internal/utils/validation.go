/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"sort"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
)

func PerformReadWrites(stub shim.ChaincodeStubInterface, rwset *kvrwset.KVRWSet) (readset [][]byte, writeset [][]byte, err error) {
	// normal reads
	var readKeys []string
	readsetMap := make(map[string][]byte)
	for _, r := range rwset.Reads {
		k := TransformToFPCKey(r.Key)
		readKeys = append(readKeys, k)
		v, _ := stub.GetState(k)
		readsetMap[k] = v

	}

	// range query reads
	for _, rqi := range rwset.RangeQueriesInfo {
		if rqi.GetRawReads() == nil {
			// no raw reads available in this range query
			continue
		}
		for _, qr := range rqi.GetRawReads().KvReads {
			k := TransformToFPCKey(qr.Key)
			readKeys = append(readKeys, k)
			v, _ := stub.GetState(k)
			readsetMap[k] = v
		}
	}

	// writes
	var writeKeys []string
	writesetMap := make(map[string][]byte)
	for _, w := range rwset.Writes {
		k := TransformToFPCKey(w.Key)
		writeKeys = append(writeKeys, k)
		writesetMap[k] = w.Value
		_ = stub.PutState(k, w.Value)
	}

	// sort readset and writeset as enclave uses a sorted map
	sort.Strings(readKeys)
	sort.Strings(writeKeys)

	// prepare sorted read/write set as output
	for _, k := range readKeys {
		readset = append(readset, []byte(k))
		readset = append(readset, readsetMap[k])
	}

	for _, k := range writeKeys {
		writeset = append(writeset, []byte(k))
		writeset = append(writeset, writesetMap[k])
	}

	return readset, writeset, nil
}
