#ifndef _TRUSTED_LEDGER_H_
#define _TRUSTED_LEDGER_H_

#include "types.h"

#ifdef __cplusplus
extern "C" {
#endif

int tlcc_create_enclave(enclave_id_t* eid, const char* enclave_file);

int tlcc_init_with_genesis(enclave_id_t eid, uint8_t* genesis, uint32_t genesis_size);

int tlcc_send_block(enclave_id_t eid, uint8_t* block, uint32_t block_size);

int tlcc_destroy_enclave(enclave_id_t eid);

uint32_t tlcc_get_quote_size(void);

int tlcc_get_local_attestation_report(
    enclave_id_t eid, target_info_t* target_info, report_t* report, ec256_public_t* pubkey);

int tlcc_get_target_info(enclave_id_t eid, target_info_t* target_info);

int tlcc_get_pk(enclave_id_t eid, ec256_public_t* pubkey);

// this is for debugging
int tlcc_print_state(enclave_id_t eid);

int tlcc_get_state_metadata(enclave_id_t eid, const char* key, uint8_t* nonce, cmac_t* cmac);

int tlcc_get_multi_state_metadata(
    enclave_id_t eid, const char* comp_key, uint8_t* nonce, cmac_t* cmac);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_TRUSTED_LEDGER_H_ */
