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

#if DO_DEBUG == true
#define LOG_DEBUG(fmt, ...) \
    printf(CYN "DEBUG " LOC_FMT TAG YEL fmt NRM "\n", __FILE__, __LINE__, ##__VA_ARGS__)
#else  // DO_DEBUG
#define LOG_DEBUG(fmt, ...)
#endif  // DO_DEBUG

#if DO_INFO == true
#define LOG_INFO(fmt, ...) \
    printf(CYN "INFO " LOC_FMT TAG NRM fmt "\n", __FILE__, __LINE__, ##__VA_ARGS__)
#else  // DO_INFO
#define LOG_INFO(fmt, ...)
#endif  // DO_INFO

#if DO_WARNING == true
#define LOG_WARNING(fmt, ...) \
    printf(CYN "WARNING " LOC_FMT TAG RED fmt NRM "\n", __FILE__, __LINE__, ##__VA_ARGS__)
#else  // DO_WARNING
#define LOG_WARNING(fmt, ...)
#endif  // DO_WARNING

#if DO_ERROR == true
#define LOG_ERROR(fmt, ...) \
    printf(CYN "ERROR " LOC_FMT TAG RED fmt NRM "\n", __FILE__, __LINE__, ##__VA_ARGS__)
#else  // DO_ERROR
#define LOG_ERROR(fmt, ...)
#endif  // DO_ERROR

#endif  // LOG_DEFINES
