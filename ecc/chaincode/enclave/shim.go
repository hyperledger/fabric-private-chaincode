//go:build !mock_ecc
// +build !mock_ecc

/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package enclave

import (
	"bytes"
	"fmt"
	"sync"
	"unsafe"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
)

// #cgo CFLAGS: -I${SRCDIR}/../../common/sgxcclib
// #include "common-sgxcclib.h"
// #include <string.h>
//
// /*
//    Below extern definitions should really be done by cgo but without we get following warning:
//       warning: implicit declaration of function ▒▒▒_GoStringPtr_▒▒▒ [-Wimplicit-function-declaration]
// */
// extern const char *_GoStringPtr(_GoString_ s);
// extern size_t _GoStringLen(_GoString_ s);
//
// static inline void _cpy_bytes(uint8_t* target, uint8_t* val, uint32_t size)
// {
//   memcpy(target, val, size);
// }
//
// static inline void _set_int(uint32_t* target, uint32_t val)
// {
//   *target = val;
// }
//
// static inline void _cpy_str(char* target, _GoString_ val, uint32_t max_size)
// {
//   #define MIN(x, y) (((x) < (y)) ? (x) : (y))
//   // Note: have to do MIN as _GoStringPtr() might not be NULL terminated.
//   // Also _GoStringLen returns the length without \0 ...
//   size_t goStrLen = _GoStringLen(val)+1;
//   snprintf(target, MIN(max_size, goStrLen), "%s", _GoStringPtr(val));
// }
//
import "C"

// Stubs is a container struct used for the callbacks
type Stubs struct {
	shimStub shim.ChaincodeStubInterface
}

// have a global registry
var registry = newRegistry()

// used to store shims for callbacks
type stubRegistry struct {
	sync.RWMutex
	index    int
	internal map[int]*Stubs
}

func newRegistry() *stubRegistry {
	return &stubRegistry{
		internal: make(map[int]*Stubs),
	}
}

func (r *stubRegistry) register(stubs *Stubs) int {
	r.Lock()
	defer r.Unlock()
	r.index++
	r.internal[r.index] = stubs
	return r.index
}

func (r *stubRegistry) release(i int) {
	r.Lock()
	delete(r.internal, i)
	r.Unlock()
}

func (r *stubRegistry) get(i int) *Stubs {
	r.RLock()
	stubs, ok := r.internal[i]
	r.RUnlock()
	if !ok {
		panic(fmt.Errorf("no shim for: %d", i))
	}
	return stubs
}

//export get_state
func get_state(key *C.char, val *C.uint8_t, max_val_len C.uint32_t, val_len *C.uint32_t, ctx unsafe.Pointer) {
	stubs := registry.get(*(*int)(ctx))

	// check if composite key
	key_str := C.GoString(key)
	if utils.IsFPCCompositeKey(key_str) {
		comp := utils.SplitFPCCompositeKey(key_str)
		key_str, _ = stubs.shimStub.CreateCompositeKey(comp[0], comp[1:])
	}

	data, err := stubs.shimStub.GetState(key_str)
	if err != nil {
		panic("error while getting state")
	}
	if C.uint32_t(len(data)) > max_val_len {
		C._set_int(val_len, C.uint32_t(0))
		// NOTE: there is currently no way to explicitly return an error
		// to distinguish from absence of key.  However, iff key exist
		// and we return an error, this should trigger an integrity
		// error, so the shim implicitly notice the difference.
		return
	}
	C._cpy_bytes(val, (*C.uint8_t)(C.CBytes(data)), C.uint32_t(len(data)))
	C._set_int(val_len, C.uint32_t(len(data)))
}

//export put_state
func put_state(key *C.char, val unsafe.Pointer, val_len C.int, ctx unsafe.Pointer) {
	stubs := registry.get(*(*int)(ctx))

	// check if composite key
	key_str := C.GoString(key)
	if utils.IsFPCCompositeKey(key_str) {
		comp := utils.SplitFPCCompositeKey(key_str)
		key_str, _ = stubs.shimStub.CreateCompositeKey(comp[0], comp[1:])
	}

	if stubs.shimStub.PutState(key_str, C.GoBytes(val, val_len)) != nil {
		panic("error while putting state")
	}
}

//export get_state_by_partial_composite_key
func get_state_by_partial_composite_key(comp_key *C.char, values *C.uint8_t, max_values_len C.uint32_t, values_len *C.uint32_t, ctx unsafe.Pointer) {
	stubs := registry.get(*(*int)(ctx))

	// split and get a proper composite key
	comp := utils.SplitFPCCompositeKey(C.GoString(comp_key))
	iter, err := stubs.shimStub.GetStateByPartialCompositeKey(comp[0], comp[1:])
	if err != nil {
		panic("error while range query")
	}
	defer iter.Close()

	var buf bytes.Buffer
	buf.WriteString("[")
	for iter.HasNext() {
		item, err := iter.Next()
		if err != nil {
			panic("Error " + err.Error())
		}
		buf.WriteString("{\"key\":\"")
		buf.WriteString(utils.TransformToFPCKey(item.Key))
		buf.WriteString("\",\"value\":\"")
		buf.Write(item.Value)
		if iter.HasNext() {
			buf.WriteString("\"},")
		} else {
			buf.WriteString("\"}")
		}
	}
	buf.WriteString("]")
	data := buf.Bytes()

	if C.uint32_t(len(data)) > max_values_len {
		C._set_int(values_len, C.uint32_t(0))
		// NOTE: there is currently no way to explicitly return an error
		// to distinguish from absence of key.  However, iff key exist
		// and we return an error, this should trigger an integrity
		// error, so the shim implicitly notice the difference.
		return
	}
	C._cpy_bytes(values, (*C.uint8_t)(C.CBytes(data)), C.uint32_t(len(data)))
	C._set_int(values_len, C.uint32_t(len(data)))
}

//export del_state
func del_state(key *C.char, ctx unsafe.Pointer) {
	stubs := registry.get(*(*int)(ctx))

	// check if composite key
	key_str := C.GoString(key)
	if utils.IsFPCCompositeKey(key_str) {
		comp := utils.SplitFPCCompositeKey(key_str)
		key_str, _ = stubs.shimStub.CreateCompositeKey(comp[0], comp[1:])
	}

	err := stubs.shimStub.DelState(key_str)
	if err != nil {
		panic("error while deleting state")
	}
}
