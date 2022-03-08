//go:build WITH_PDO_CRYPTO
// +build WITH_PDO_CRYPTO

/*
   Copyright 2019 Intel Corporation
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package pdo

// extern void golog(char*);
//
// int golog_cgo_wrapper(const char* str)
// {
//      golog((char*)str);
//      return 1;
// }
//
import "C"
