/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef logging_h
#define logging_h

#include <stdbool.h>
#include <stdio.h>  // for printf
#include "../include/log-defines.h"

/*
 * log_callback_f is the function type that the user must pass as a callback,
 * through `logging_set_callback`.
 * The expected return value roughly follows the standard definitions of `printf` and `puts`:
 * - negative int for error
 * - non-negative for success
 */
typedef int (*log_callback_f)(const char* str);

#ifdef __cplusplus
extern "C" {
#endif

/*
 * `logging_set_callback` lets a user set the callback logging function.
 * By default, no callback function is set.
 * So the `loggingf` function below fails if no callback is initialized.
 */
bool logging_set_callback(log_callback_f log_callback);

/*
 * `loggingf` forwards the input string to the initialized callback function.
 * By default, no callback function is initialized, so `loggingf` returns error.
 * Returns a boolean as integer:
 * 0 false/error
 * >0 true/success
 *
 * The function prototype is in "../include/log-defines.h"
 */

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif  // logging_h
