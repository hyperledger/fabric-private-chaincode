/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include <assert.h>
#include <stdio.h>
#include <string.h>

#include <glob.h>

#include "error.h"
#include "logging.h"

#include "common-sgxcclib.h"
#include "trusted_ledger.h"

#include "sgx_quote.h"

#define ENCLAVE_FILENAME "enclave.signed.so"

void golog(const char* fmt, ...)
{
    char buf[BUFSIZ] = {'\0'};
    va_list ap;
    va_start(ap, fmt);
    vsnprintf(buf, BUFSIZ, fmt, ap);
    va_end(ap);
    printf("%s", buf);
}

int failNow()
{
    exit(-1);
}

int read_block(const char* filename, uint8_t** out)
{
    // load genesis blob
    uint32_t len = 0;
    FILE* fp = fopen(filename, "r");
    if (fp != NULL)
    {
        if (fseek(fp, 0L, SEEK_END) == 0)
        {
            long bufsize = ftell(fp);
            if (bufsize == -1)
            {
                LOG_ERROR("Test: buffersize == -1");
                failNow();
            }

            *out = malloc(sizeof(uint8_t) * (bufsize));

            if (fseek(fp, 0L, SEEK_SET) != 0)
            {
                LOG_ERROR("Test: fseek != 0");
                failNow();
            }

            len = fread(*out, sizeof(uint8_t), bufsize, fp);
            if (ferror(fp) != 0)
            {
                LOG_ERROR("Test: Error reading file");
            }
        }
        fclose(fp);
    }
    else
    {
        LOG_ERROR("Test: Can not open block file %s", filename);
        failNow();
    }
    return len;
}

int test_get_and_verify_cmac(enclave_id_t eid)
{
    // first get some non-existing data
    LOG_INFO("Test: Verify non-existing state");
    const char* key_not_exists = "this.does.not.exist";
    uint8_t nonce[32] = {123};
    cmac_t cmac = {0};

    int ret = tlcc_get_state_metadata(eid, key_not_exists, nonce, &cmac);
    if (ret != 0)
    {
        LOG_ERROR("Test: ERROR - get state: %d", ret);
        failNow();
    }

    // second with exising data
    LOG_INFO("Test: Get state");
    const char* key = "ecc.somePrefx.MyAuction.Johan";

    // first get some data
    ret = tlcc_get_state_metadata(eid, key, nonce, &cmac);
    if (ret != 0)
    {
        LOG_ERROR("Test: ERROR - get state: %d", ret);
        failNow();
    }

    LOG_INFO("Test: Verify range query");
    const char* comp_key = "ecc.";
    // first get some data

    int num_of_runs = 5;
    for (int i = 0; i < num_of_runs; i++)
    {
        LOG_INFO("Test: Invoke #%d", i);

        ret = tlcc_get_multi_state_metadata(eid, comp_key, nonce, &cmac);
        if (ret != 0)
        {
            LOG_ERROR("Test: ERROR - get state: %d", ret);
            failNow();
        }
    }

    return 0;
}

/* Application entry */
int main(int argc, char** argv)
{
    (void)(argc);
    (void)(argv);

    LOG_INFO("Test: Create enclave and send blocks");
    enclave_id_t eid;
    if (sgxcc_create_enclave(&eid, ENCLAVE_FILENAME) < 0)
    {
        LOG_ERROR("Test: Can not create enclave!!!");
        failNow();
    }

    // test blocks
    glob_t globbed_test_files = {0, NULL, 0};
    glob("test/test_blocks/*-block[1-9]", GLOB_DOOFFS, NULL, &globbed_test_files);
    glob(
        "test/test_blocks/*-block[1-9][0-9]", GLOB_DOOFFS | GLOB_APPEND, NULL, &globbed_test_files);
    LOG_INFO("Test: Found %ld non-genesis blocks", globbed_test_files.gl_pathc);

    // init enclave with genesis block
    char* genesis_block_name = "test/test_blocks/mychannel-block0";
    uint8_t* genesis = NULL;
    int genesis_size = read_block(genesis_block_name, &genesis);
    LOG_INFO("Test: Send genesis block: \"%s\"", genesis_block_name);

    int numOfIterations = 1;
    for (int i = 0; i < numOfIterations; i++)
    {
        LOG_INFO("Test: #%d", i);
        tlcc_init_with_genesis(eid, genesis, genesis_size);

        // send blocks
        for (int j = 0; j < globbed_test_files.gl_pathc; j++)
        {
            uint8_t* block = NULL;
            int block_size = read_block(globbed_test_files.gl_pathv[j], &block);
            LOG_INFO("Test: Send block: \"%s\"", globbed_test_files.gl_pathv[j]);
            tlcc_send_block(eid, block, block_size);
            free(block);
        }
    }

    // show what we got
    tlcc_print_state(eid);
    test_get_and_verify_cmac(eid);
    sgxcc_destroy_enclave(eid);
    free(genesis);
    globfree(&globbed_test_files);

    return 0;
}
