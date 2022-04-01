//go:build !mock_ecc
// +build !mock_ecc

/*
   Copyright 2019 Intel Corporation
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package enclave

import (
	"github.com/hyperledger/fabric/common/flogging"
)

// #cgo CFLAGS: -I${SRCDIR}/../../../common/logging/untrusted
// #include "logging.h"
//
// extern int golog_cgo_wrapper(const char* str);
//
import "C"

var enclaveLogger = flogging.MustGetLogger("ecc_enclave")

func init() {
	r := C.logging_set_callback(C.log_callback_f(C.golog_cgo_wrapper))
	if !r {
		panic("error initializing logging for cgo")
	}
}

//export golog
func golog(str *C.char) {
	enclaveLogger.Infof("%s", C.GoString(str))
}
