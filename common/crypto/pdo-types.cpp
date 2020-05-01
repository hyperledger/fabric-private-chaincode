/* Copyright 2018 Intel Corporation
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

/* NOTE ***********************************************************************
 *
 * This file has been copied from the Private Data Objects repo and slightly
 * modified. This file provides types (and type casts), including encoding and
 * decoding of base64 strings. As PDO uses different function signatures for
 * base64 features with respect to FPC, the respective function calls have been
 * updated.
 *
 * Recommendation for future improvement: import PDO's base64 package in FPC.
 * 
 * PDO has modified the base64 package for performance reasons -- namely to
 * reduce copies of data in memory.
 * So by importing the package, FPC can:
 *  - benefit from the performance improvements,
 *  - and drop this modified `types.cpp` file.
 *
 * ****************************************************************************
 */

#include <algorithm>
#include <string>
#include <vector>

#include "types.h"
#include "base64.h"
#include "hex_string.h"

// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
// Simple conversion from ByteArray to std::string
// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
std::string ByteArrayToString(const ByteArray& inArray)
{
    std::string outString;
    std::transform(inArray.begin(), inArray.end(), std::back_inserter(outString),
                   [](unsigned char c) -> char { return (char)c; });

    return outString;
}

// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
// Conversion from byte array to string array
// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
void ByteArrayToStringArray(const ByteArray& inByteArray, StringArray& outStringArray)
{
    outStringArray.resize(0);
    std::transform(inByteArray.begin(), inByteArray.end(), std::back_inserter(outStringArray),
                   [](unsigned char c) -> char { return (char)c; });
}

StringArray ByteArrayToStringArray(const ByteArray& inByteArray)
{
    StringArray outStringArray(0);
    ByteArrayToStringArray(inByteArray, outStringArray);
    return outStringArray;
}

// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
// Simple conversion from ByteArray to Base64EncodedString
// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
Base64EncodedString ByteArrayToBase64EncodedString(const ByteArray& buf)
{
    return base64_encode(buf.data(), buf.size());
}  // ByteArrayToBase64EncodedString

// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
// Simple conversion from Base64EncodedString to ByteArray
ByteArray Base64EncodedStringToByteArray(const Base64EncodedString& encoded)
{
    std::string s = base64_decode(encoded);
    ByteArray b;
    std::transform(s.begin(), s.end(), std::back_inserter(b),
                   [](unsigned char c) -> char { return (uint8_t)c; });
    return b;
}  // Base64EncodedStringToByteArray

// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
// Simple conversion from ByteArray to HexEncodedString
HexEncodedString ByteArrayToHexEncodedString(const ByteArray& buf)
{
    return pdo::BinaryToHexString(buf);
}  // ByteArrayToHexEncodedString

// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
// Simple conversion from HexEncodedString to ByteArray
// throws ValueError
ByteArray HexEncodedStringToByteArray(const HexEncodedString& encoded)
{
    return pdo::HexStringToBinary(encoded);
}  // HexEncodedStringToByteArray
