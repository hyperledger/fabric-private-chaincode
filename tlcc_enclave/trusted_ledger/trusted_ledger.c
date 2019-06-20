#include "trusted_ledger.h"
#include "enclave_u.h"

#include <stdbool.h>
#include <string.h>
#include <unistd.h>

#include "sgx_eid.h"  // sgx_enclave_id_t
#include "sgx_quote.h"
#include "sgx_uae_service.h"
#include "sgx_urts.h"

#define NRM "\x1B[0m"
#define RED "\x1B[31m"
#define CYN "\x1B[36m"

#define PERR(fmt, ...) golog(CYN "ERROR" RED fmt NRM "\n", ##__VA_ARGS__)

// extern go printf
extern void golog(const char* format, ...);

int tlcc_create_enclave(enclave_id_t* eid, const char* enclave_file)
{
    sgx_launch_token_t token = {0};
    int updated = 0;

    if (access(enclave_file, F_OK) == -1)
    {
        PERR("Lib: enclave file does not exist! %s", enclave_file);
        return -1;
    }

    int ret = sgx_create_enclave(enclave_file, SGX_DEBUG_FLAG, &token, &updated, eid, NULL);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: Unable to create enclave. reason: %d", ret);
        return -1;
    }

    return 0;
}

int tlcc_destroy_enclave(enclave_id_t eid)
{
    int ret = sgx_destroy_enclave((sgx_enclave_id_t)eid);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: Error: %s", ret);
        return ret;
    }

    return SGX_SUCCESS;
}

uint32_t tlcc_get_quote_size()
{
    uint32_t needed_quote_size = 0;
    sgx_calc_quote_size(NULL, 0, &needed_quote_size);
    return needed_quote_size;
}

int tlcc_get_target_info(enclave_id_t eid, target_info_t* target_info)
{
    int enclave_ret = -1;
    int ret = ecall_get_target_info(eid, &enclave_ret, (sgx_target_info_t*)target_info);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: ERROR - ecall_target_info: %s", ret);
    }

    return ret;
}

int tlcc_get_local_attestation_report(
    enclave_id_t eid, target_info_t* target_info, report_t* report, ec256_public_t* pubkey)
{
    int enclave_ret = -1;
    int ret = ecall_create_report(eid, &enclave_ret, (sgx_target_info_t*)target_info,
        (sgx_report_t*)report, (uint8_t*)pubkey);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: ERROR - ecall_create_report: %d", ret);
        return ret;
    }

    return ret;
}

int tlcc_get_pk(enclave_id_t eid, ec256_public_t* pubkey)
{
    int enclave_ret = -1;
    int ret = ecall_get_pk(eid, &enclave_ret, (uint8_t*)pubkey);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: ERROR - invoke: %d", ret);
        return ret;
    }

    return enclave_ret;
}

int tlcc_init_with_genesis(enclave_id_t eid, uint8_t* genesis, uint32_t genesis_size)
{
    int enclave_ret = -1;
    int ret = ecall_init(eid, &enclave_ret);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        PERR("Lib: Unable to initialize enclave. reason: %d %d", ret, enclave_ret);
        return -1;
    }

    ret = ecall_join_channel(eid, &enclave_ret, genesis, genesis_size);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        PERR("Lib: Unable to join channel. reason: %d %d", ret, enclave_ret);
        return -1;
    }

    return enclave_ret;
}

int tlcc_send_block(enclave_id_t eid, uint8_t* block, uint32_t block_size)
{
    int enclave_ret = -1;
    int ret = ecall_next_block(eid, &enclave_ret, block, block_size);
    if (ret != SGX_SUCCESS || enclave_ret != SGX_SUCCESS)
    {
        PERR("Lib: ERROR Process block within enclave. reason: %d %d", ret, enclave_ret);
        return ret;
    }

    return enclave_ret;
}

int tlcc_get_state_metadata(enclave_id_t eid, const char* key, uint8_t* nonce, cmac_t* cmac)
{
    int enclave_ret = -1;
    int ret = ecall_get_state_metadata(eid, (int*)&enclave_ret, key, nonce, cmac);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: Error: %d", ret);
        return ret;
    }

    return SGX_SUCCESS;
}

int tlcc_get_multi_state_metadata(
    enclave_id_t eid, const char* comp_key, uint8_t* nonce, cmac_t* cmac)
{
    int enclave_ret = -1;
    int ret = ecall_get_multi_state_metadata(eid, (int*)&enclave_ret, comp_key, nonce, cmac);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: Error: %d", ret);
        return ret;
    }

    return SGX_SUCCESS;
}

// this is only for debugging
int tlcc_print_state(enclave_id_t eid)
{
    int enclave_ret = -1;
    int ret = ecall_print_state(eid, (int*)&enclave_ret);
    if (ret != SGX_SUCCESS)
    {
        PERR("Lib: Error: %d", ret);
        return ret;
    }

    return SGX_SUCCESS;
}

/* OCall functions */
void ocall_print_string(const char* str)
{
    golog(str);
}
