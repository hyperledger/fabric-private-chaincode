//go:build WITH_PDO_CRYPTO
// +build WITH_PDO_CRYPTO

/*
   Copyright 2019 Intel Corporation
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package pdo

import (
	"github.com/hyperledger/fabric/common/flogging"
)

// #cgo CFLAGS: -I${SRCDIR}/../../../../common/logging/untrusted
// #cgo LDFLAGS: -L${SRCDIR}/../../../../common/logging/_build -lulogging
// #include "logging.h"
//
// extern int golog_cgo_wrapper(const char* str);
//
import "C"

var logger = flogging.MustGetLogger("cgo")

func init() {
	logger.Info("Initializing logger")
	r := C.logging_set_callback(C.log_callback_f(C.golog_cgo_wrapper))
	if !r {
		panic("error initializing logging for cgo")
	}
}

//export golog
func golog(str *C.char) {
	logger.Infof("%s", C.GoString(str))
}
