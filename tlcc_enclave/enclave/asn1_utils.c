/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "asn1_utils.h"

ASN1_SEQUENCE(ASN1BlockHeader) = {ASN1_SIMPLE(ASN1BlockHeader, number, ASN1_INTEGER),
    ASN1_SIMPLE(ASN1BlockHeader, prev_hash, ASN1_OCTET_STRING),
    ASN1_SIMPLE(ASN1BlockHeader, data_hash, ASN1_OCTET_STRING)} ASN1_SEQUENCE_END(ASN1BlockHeader)

    IMPLEMENT_ASN1_FUNCTIONS(ASN1BlockHeader)

        uint32_t block_header2DER(common_BlockHeader * header, uint8_t** DERHeader)
{
    ASN1BlockHeader* asn1BlockHeader = ASN1BlockHeader_new();

    // block number
    ASN1_INTEGER_set(asn1BlockHeader->number, header->number);

    // NOTE for genesis block prev_hash is empty and does not need to be set;
    // however, if we would do it anyway
    // the asn1 block header would contain 32 zeros and would be incorrect
    if (header->number > 0)
    {
        // prev hash
        ASN1_STRING_set(asn1BlockHeader->prev_hash, &header->previous_hash, 32);
    }

    // data hash
    ASN1_STRING_set(asn1BlockHeader->data_hash, &header->data_hash, 32);

    // create DER bytes
    uint32_t DERHeader_len = i2d_ASN1BlockHeader(asn1BlockHeader, NULL);
    *DERHeader = malloc(DERHeader_len);
    uint8_t* tmp = *DERHeader;
    i2d_ASN1BlockHeader(asn1BlockHeader, &tmp);

    // free asn1 header
    ASN1BlockHeader_free(asn1BlockHeader);
    return DERHeader_len;
}
