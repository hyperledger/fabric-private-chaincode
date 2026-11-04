/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <string>

typedef enum
{
    EC_UNDEFINED,              // 0
    EC_SUCCESS,                // 1
    EC_ERROR,                  // 2
    EC_HIDDEN,                 // 3
    EC_BAD_FUNCTION_NAME,      // 4
    EC_INVALID_INPUT,          // 5
    EC_MEMORY_ERROR,           // 6
    EC_SHORT_RESPONSE_BUFFER,  // 7
    EC_BAD_PARAMETERS,         // 8
} error_codes_e;

namespace Contract
{
class ErrorReport
{
private:
    error_codes_e ec_;
    std::string errorString_;

public:
    ErrorReport();

    void toStatusProtoString(std::string& outputString);
    void set(error_codes_e ec, const std::string& errorString);
    bool isSuccess();
};
}  // namespace Contract

#define CUSTOM_ERROR_REPORT(er, code, message) er.set(code, std::string(#code) + ":" + message);

#define DEFAULT_ERROR_REPORT(er, code) \
    er.set(code, std::string(#code) + ":" + std::string(__FILE__) + ":" + std::to_string(__LINE__));

#define FAST_FAIL_CHECK(errorReport, code, b)       \
    {                                               \
        if (b)                                      \
        {                                           \
            DEFAULT_ERROR_REPORT(errorReport, code) \
            return false;                           \
        }                                           \
    }

#define FAST_FAIL_CHECK_EX(parentErrorReport, pChildErrorReport, code, b) \
    {                                                                     \
        if (b)                                                            \
        {                                                                 \
            if (pChildErrorReport)                                        \
            {                                                             \
                parentErrorReport = *pChildErrorReport;                   \
            }                                                             \
            else                                                          \
            {                                                             \
                DEFAULT_ERROR_REPORT(parentErrorReport, code)             \
            }                                                             \
            return false;                                                 \
        }                                                                 \
    }

#define CATCH(b, expr) \
    do                 \
    {                  \
        try            \
        {              \
            expr;      \
            b = true;  \
        }              \
        catch (...)    \
        {              \
            b = false; \
        }              \
    } while (0);
