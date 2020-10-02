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

#define LOG_DEBUG(fmt, ...) \
    if (DO_DEBUG)           \
    printf(CYN "DEBUG " LOC_FMT TAG YEL fmt NRM "\n", __FILE__, __LINE__, ##__VA_ARGS__)

#define LOG_INFO(fmt, ...) \
    if (DO_INFO)           \
    printf(CYN "INFO " LOC_FMT TAG NRM fmt "\n", __FILE__, __LINE__, ##__VA_ARGS__)

#define LOG_WARNING(fmt, ...) \
    if (DO_WARNING)           \
    printf(CYN "WARNING " LOC_FMT TAG RED fmt NRM "\n", __FILE__, __LINE__, ##__VA_ARGS__)

#define LOG_ERROR(fmt, ...) \
    if (DO_ERROR)           \
    printf(CYN "ERROR " LOC_FMT TAG RED fmt NRM "\n", __FILE__, __LINE__, ##__VA_ARGS__)

#endif  // LOG_DEFINES
