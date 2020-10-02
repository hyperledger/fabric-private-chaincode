include(${CMAKE_MODULE_PATH}/NanoPB.cmake)
include(${FPC_PATH}/tlcc_enclave/enclave/CMakeVariables.txt)

set(SOURCE_FILES
    asn1_utils.c
    crypto.cpp
    ecc_json.cpp
    enclave.cpp
    enclave_t.c
    ias.cpp
    ledger.cpp
)

set(COMMON_SOURCE_FILES
    ${COMMON_SOURCE_DIR}/enclave/common.cpp
    ${COMMON_SOURCE_DIR}/base64/base64.cpp
    ${COMMON_SOURCE_DIR}/json/parson.c
    ${COMMON_SOURCE_DIR}/utils.c
    ${COMMON_SOURCE_DIR}/logging.c
    ${NANOPB_PATH}/pb_common.c
    ${NANOPB_PATH}/pb_decode.c
    ${NANOPB_PATH}/pb_encode.c
)

set(PROTO_SOURCE_FILES
    ${COMMON_SOURCE_DIR}/protos/fabric/common/common.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/common/configtx.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/common/configuration.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/common/ledger.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/common/policies.pb.c
    ${COMMON_SOURCE_DIR}/protos/google/protobuf/empty.pb.c
    ${COMMON_SOURCE_DIR}/protos/google/protobuf/timestamp.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/ledger/queryresult/kv_query_result.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/ledger/rwset/kvrwset/kv_rwset.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/ledger/rwset/rwset.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/msp/identities.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/msp/msp_config.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/msp/msp_principal.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/chaincode.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/chaincode_event.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/chaincode_shim.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/configuration.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/events.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/peer.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/proposal.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/proposal_response.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/query.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/signed_cc_dep_spec.pb.c
    ${COMMON_SOURCE_DIR}/protos/fabric/peer/transaction.pb.c
)

add_definitions(-DENCLAVE_CODE)
add_definitions(-DPB_ENABLE_MALLOC)
add_definitions(-DPB_FIELD_32BIT)

set(cleanup_files
    ${CMAKE_CURRENT_SOURCE_DIR}/enclave_t.c
    ${CMAKE_CURRENT_SOURCE_DIR}/enclave_t.h
    ${CMAKE_LIBRARY_OUTPUT_DIRECTORY}/enclave.signed.so
    ${CMAKE_LIBRARY_OUTPUT_DIRECTORY}/mrenclave
    ${CMAKE_LIBRARY_OUTPUT_DIRECTORY}/mrenclave.go)
set_directory_properties(PROPERTIES ADDITIONAL_MAKE_CLEAN_FILES "${cleanup_files}")

add_custom_command(
    OUTPUT ${CMAKE_CURRENT_SOURCE_DIR}/enclave_t.c
    COMMAND ${SGX_EDGER8R} --trusted enclave.edl
    --search-path ${COMMON_SOURCE_DIR}/enclave
    --search-path ${COMMON_SOURCE_DIR}/crypto
    --search-path ${COMMON_SOURCE_DIR}/logging/ocalls
    --search-path ${SGX_SDK}/include
    --search-path ${SGX_SSL}/include
    COMMENT "Generating enclave_t.{c,h}"
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    )

add_library(enclave SHARED ${SOURCE_FILES} ${COMMON_SOURCE_FILES} ${PROTO_SOURCE_FILES})

include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${COMMON_SOURCE_DIR}/protos
    ${COMMON_SOURCE_DIR}/protos/fabric
    ${COMMON_SOURCE_DIR}
    ${COMMON_SOURCE_DIR}/base64
    ${COMMON_SOURCE_DIR}/enclave
    ${COMMON_SOURCE_DIR}/json
    ${NANOPB_PATH}
    ${SGX_SDK}/include
    ${SGX_SDK}/include/tlibc
    ${SGX_SDK}/include/libcxx
    ${SGX_SSL}/include
    ${FPC_PATH}/common/crypto
    ${FPC_PATH}/common/crypto/pdo/common
    )

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${SGX_COMMON_CFLAGS} -nostdinc -fno-builtin -fvisibility=hidden -fpie -fstack-protector -std=c11 -include tsgxsslio.h")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${SGX_COMMON_CFLAGS} -nostdinc -fno-builtin -fvisibility=hidden -fpie -fstack-protector -std=c++11 -nostdinc++ -include tsgxsslio.h")

set_target_properties(enclave PROPERTIES PREFIX "")

target_link_libraries(enclave "${SGX_COMMON_CFLAGS} -Wl,--no-undefined -nostdlib -nodefaultlibs -nostartfiles \
-L${COMMON_SOURCE_DIR}/crypto/_build \
-Wl,--start-group -l${T_CRYPTO_LIB_NAME} -l${T_CRYPTO_ADAPT_LIB_NAME} -Wl,--end-group \
-L${SGX_SSL_LIBRARY_PATH} \
-Wl,--whole-archive -lsgx_tsgxssl -Wl,--no-whole-archive -lsgx_tsgxssl_crypto \
-L${SGX_LIBRARY_PATH} \
-Wl,--whole-archive -l${SGX_TRTS_LIB} -Wl,--no-whole-archive \
-Wl,--start-group -lsgx_tstdc -lsgx_tcxx -lsgx_tcrypto -l${SGX_TSVC_LIB} -Wl,--end-group \
-Wl,-Bstatic -Wl,-Bsymbolic -Wl,--no-undefined \
-Wl,-pie,-eenclave_entry -Wl,--export-dynamic \
-Wl,--defsym,__ImageBase=0 \
-Wl,--version-script=${CMAKE_CURRENT_SOURCE_DIR}/enclave.lds")

add_custom_command(TARGET enclave
    POST_BUILD
    COMMENT "Create enclave signer's key"
    COMMAND openssl genrsa -3 -out ${CMAKE_BINARY_DIR}/enclave_signer_private.pem 3072
    COMMENT "Signing the enclave => ${CMAKE_LIBRARY_OUTPUT_DIRECTORY}/enclave.signed.so"
    COMMAND mkdir -p ${CMAKE_BINARY_DIR}/lib
    COMMAND ${SGX_ENCLAVE_SIGNER} sign
    -key ${CMAKE_BINARY_DIR}/enclave_signer_private.pem
    -config ${CMAKE_CURRENT_SOURCE_DIR}/enclave.config.xml
    -enclave enclave.so
    -out ${CMAKE_BINARY_DIR}/enclave.signed.so
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR})

add_custom_command(TARGET enclave
    POST_BUILD
    COMMAND ${COMMON_SOURCE_DIR}/enclave/generate_mrenclave.sh ${CMAKE_BINARY_DIR} ${CMAKE_CURRENT_SOURCE_DIR}
    COMMAND cp ${CMAKE_BINARY_DIR}/mrenclave ${CMAKE_BINARY_DIR}/lib
    COMMAND cp ${CMAKE_BINARY_DIR}/enclave.signed.so ${CMAKE_BINARY_DIR}/lib
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
    COMMENT "Create enclave mrenclave")
