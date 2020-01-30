/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <string>

typedef enum
{
    EC_UNDEFINED,                 // 0
    EC_SUCCESS,                   // 1
    EC_ERROR,                     // 2
    EC_HIDDEN,                    // 3
    EC_BAD_FUNCTION_NAME,         // 4
    EC_INVALID_INPUT,             // 5
    EC_MEMORY_ERROR,              // 6
    EC_SHORT_RESPONSE_BUFFER,     // 7
    EC_BAD_PARAMETERS,            // 8
    EC_AUCTION_RELATED_CODES,     // 9
    EC_ROUND_ACTIVE,              // 10
    EC_RESTRICTED_AUCTION_STATE,  // 11
    EC_NOT_IN_CLOCK_PHASE,        // 12
    EC_NOT_IN_ASSIGNMENT_PHASE,   // 13
    EC_ROUND_NOT_CURRENT,         // 14
    EC_UNRECOGNIZED_TERRITORY,    // 15
    EC_UNAVAILABLE_QUANTITY,      // 16
    EC_DUPLICATE_TERRITORIES,     // 17
    EC_PRICE_OUT_OF_RANGE,        // 18
    EC_ROUND_NOT_ACTIVE,          // 19
    EC_NOT_ENOUGH_ELIGIBILITY,    // 20
    EC_TOO_MUCH_DEMAND,           // 21
    EC_UNRECOGNIZED_SUBMITTER,    // 22
    EC_BELOW_POSTED_PRICE,        // 23
    EC_ABOVE_CLOCK_PRICE,         // 24
    EC_EVALUATION_ERROR,          // 25
    EC_UNIMPLEMENTED_API          // 26
} error_codes_e;

namespace ClockAuction
{
class ErrorReport
{
private:
    error_codes_e ec_;
    std::string errorString_;

public:
    ErrorReport();

    void set(error_codes_e ec, const std::string& errorString);
    void toStatusJsonString(std::string& jsonString);
    void toWrappedStatusJsonString(std::string& jsonString);
    bool isSuccess();
};
}  // namespace ClockAuction

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
