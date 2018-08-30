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

#ifndef asn1_util_h
#define asn1_util_h

#include "common/common.pb.h"
#include "openssl/asn1.h"
#include "openssl/asn1t.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    ASN1_INTEGER *number;
    ASN1_OCTET_STRING *prev_hash;
    ASN1_OCTET_STRING *data_hash;
} ASN1BlockHeader;

DECLARE_ASN1_FUNCTIONS(ASN1BlockHeader)

uint32_t block_header2DER(common_BlockHeader *header, uint8_t **DERHeader);

#ifdef __cplusplus
}
#endif

#endif /* asn1_util_h */
