/*
 * Copyright IBM Corp. All Rights Reserved.
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef LOG_DEFINES
#define LOG_DEFINES

#ifndef TAG
#define TAG ""
#endif

#define LOC_FMT " (%s:%d) "

#define NRM "\x1B[0m"
#define CYN "\x1B[36m"
#define YEL "\x1B[33m"
#define RED "\x1B[31m"

/*
 * Note: `DO_DEBUG` is set to `false` by default, so no `LOG_DEBUG` is displayed.
 * At compile time, this behaviour can be changed by defining `-DDO_DEBUG=true` before the header is
 * included. In SGX deployments, such define should be set "only" when the `SGX_BUILD` environment
 * variable is set to `DEBUG`. Finally, notice that `DO_INFO`, `DO_WARNING` and `DO_ERROR` are set
 * to `true` by default. So, unless they are explictly disabled at compile time, the respective logs
 * will be displayed.
 */

#ifndef DO_DEBUG
#define DO_DEBUG false
#endif

#ifndef DO_INFO
#define DO_INFO true
#endif

#ifndef DO_WARNING
#define DO_WARNING true
#endif

#ifndef DO_ERROR
#define DO_ERROR true
#endif

#ifdef __cplusplus
extern "C" {
#endif
int loggingf(const char* fmt, ...);
#ifdef __cplusplus
}
#endif

#if DO_DEBUG == true
#define LOG_DEBUG(fmt, ...) \
    loggingf(CYN "DEBUG " LOC_FMT TAG YEL fmt NRM "\n", __FILE__, __LINE__, ##__VA_ARGS__)
#else  // DO_DEBUG
#define LOG_DEBUG(fmt, ...)
#endif  // DO_DEBUG

#if DO_INFO == true
#define LOG_INFO(fmt, ...) \
    loggingf(CYN "INFO " LOC_FMT TAG NRM fmt "\n", __FILE__, __LINE__, ##__VA_ARGS__)
#else  // DO_INFO
#define LOG_INFO(fmt, ...)
#endif  // DO_INFO

#if DO_WARNING == true
#define LOG_WARNING(fmt, ...) \
    loggingf(CYN "WARNING " LOC_FMT TAG RED fmt NRM "\n", __FILE__, __LINE__, ##__VA_ARGS__)
#else  // DO_WARNING
#define LOG_WARNING(fmt, ...)
#endif  // DO_WARNING

#if DO_ERROR == true
#define LOG_ERROR(fmt, ...) \
    loggingf(CYN "ERROR " LOC_FMT TAG RED fmt NRM "\n", __FILE__, __LINE__, ##__VA_ARGS__)
#else  // DO_ERROR
#define LOG_ERROR(fmt, ...)
#endif  // DO_ERROR

#define ERROR_LOG_STRING "error log - omitted"

#endif  // LOG_DEFINES
