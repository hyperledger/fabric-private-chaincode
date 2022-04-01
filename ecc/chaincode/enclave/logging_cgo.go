//go:build !mock_ecc
// +build !mock_ecc

/*
   Copyright 2019 Intel Corporation
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package enclave

// extern void golog(char*);
//
// int golog_cgo_wrapper(const char* str)
// {
//      golog((char*)str);
//      return 1;
// }
//
import "C"
