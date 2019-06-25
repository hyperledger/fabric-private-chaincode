/* Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "sgx_quote.h"

#ifdef USE_EPID_LINKABLE

#define SGX_QUOTE_SIGN_TYPE SGX_LINKABLE_SIGNATURE

#endif

#ifdef USE_EPID_UNLINKABLE

#define SGX_QUOTE_SIGN_TYPE SGX_UNLINKABLE_SIGNATURE

#endif
