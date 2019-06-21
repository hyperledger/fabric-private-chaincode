/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef asn1_util_h
#define asn1_util_h

#include "common/common.pb.h"
#include "openssl/asn1.h"
#include "openssl/asn1t.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct
{
    ASN1_INTEGER* number;
    ASN1_OCTET_STRING* prev_hash;
    ASN1_OCTET_STRING* data_hash;
} ASN1BlockHeader;

DECLARE_ASN1_FUNCTIONS(ASN1BlockHeader)

uint32_t block_header2DER(common_BlockHeader* header, uint8_t** DERHeader);

#ifdef __cplusplus
}
#endif

#endif /* asn1_util_h */
