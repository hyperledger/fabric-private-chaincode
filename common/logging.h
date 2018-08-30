/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

#ifndef logging_h
#define logging_h

#include <stdbool.h>

#ifdef ENCLAVE_CODE
#define TAG "[Enclave] "
#ifdef __cplusplus
extern "C" {
#endif
int printf(const char* format, ...);
#ifdef __cplusplus
}
#endif
#else
#include <stdio.h>  // for printf
#endif

#ifndef TAG
#define TAG ""
#endif

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

#ifndef DO_ERROR
#define DO_ERROR true
#endif

#define LOG_DEBUG(fmt, ...) \
    if (DO_DEBUG) printf(CYN "DEBUG " TAG YEL fmt NRM "\n", ##__VA_ARGS__)

#define LOG_INFO(fmt, ...) \
    if (DO_INFO) printf(CYN "INFO " TAG NRM fmt "\n", ##__VA_ARGS__)

#define LOG_ERROR(fmt, ...) \
    if (DO_ERROR) printf(CYN "ERROR " TAG RED fmt NRM "\n", ##__VA_ARGS__)

#endif
