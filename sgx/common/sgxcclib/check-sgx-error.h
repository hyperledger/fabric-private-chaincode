/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef _CHECK_SGX_ERROR_H_
#define _CHECK_SGX_ERROR_H_

#include "logging.h"

#define CHECK_SGX_ERROR_AND_RETURN_ON_ERROR(sgx_status_ret)                                        \
    if (sgx_status_ret != SGX_SUCCESS)                                                             \
    {                                                                                              \
        LOG_ERROR(                                                                                 \
            "Lib: ERROR - %s:%d: " #sgx_status_ret "=%d", __FUNCTION__, __LINE__, sgx_status_ret); \
        return sgx_status_ret;                                                                     \
    }

#endif /* _CHECK_SGX_ERROR_H_ */
