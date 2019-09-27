SET(FPC_ATTESTATION_TYPE "$ENV{FPC_ATTESTATION_TYPE}")

IF("$ENV{FPC_ATTESTATION_TYPE} " STREQUAL " ")
    #default attestation flag is "epid_unlinkable"
    SET(SGX_ATTESTATION_FLAG "-DUSE_EPID_UNLINKABLE")
ELSE()
    IF("${FPC_ATTESTATION_TYPE} " STREQUAL "epid_linkable ")
        SET(SGX_ATTESTATION_FLAG "-DUSE_EPID_LINKABLE")
    ELSEIF("${FPC_ATTESTATION_TYPE} " STREQUAL "epid_unlinkable ")
        SET(SGX_ATTESTATION_FLAG "-DUSE_EPID_UNLINKABLE")
    ELSE()
        message(FATAL_ERROR "if set, FPC_ATTESTATION_TYPE must be 'epid_linkable' or 'epid_unlinkable'")
    ENDIF()
ENDIF()
message(WARNING "TLCC attestation type derived from ECC -- TLCC does not use remote attestation")

set(SOURCE_FILES
    test.c
    ${COMMON_SOURCE_DIR}/base64/base64.cpp
    ${COMMON_SOURCE_DIR}/utils.c
    ${COMMON_SOURCE_DIR}/sgxcclib/common-sgxcclib.c
    )

set(TEST_NAME test_runner)

add_executable(${TEST_NAME} ${SOURCE_FILES})

include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${CMAKE_CURRENT_SOURCE_DIR}/../trusted_ledger
    ${COMMON_SOURCE_DIR}
    ${COMMON_SOURCE_DIR}/base64
    ${COMMON_SOURCE_DIR}/sgxcclib
    ${SGX_SDK}/include
    )

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${SGX_COMMON_CFLAGS} -fPIC -Wall -Wno-attributes -std=c11 ${SGX_ATTESTATION_FLAG}")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${SGX_COMMON_CFLAGS} -fpie -std=c++11 -fpermissive")

target_link_libraries(${TEST_NAME}
    -L${SGX_LIBRARY_PATH} -l${SGX_URTS_LIB} -l${SGX_USVC_LIB} -L${SGX_SSL_LIBRARY_PATH} -lsgx_usgxssl 
    -L${CMAKE_BINARY_DIR} -ltl -pthread -lcurl -lcrypto)

add_custom_command(TARGET ${TEST_NAME}
    POST_BUILD
    COMMAND cp -rv ${CMAKE_CURRENT_SOURCE_DIR}/test_blocks ${CMAKE_BINARY_DIR}/test
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
    COMMENT "Copy test blocks")


add_dependencies(${TEST_NAME} libtl)

add_test(NAME ${TEST_NAME} 
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
    COMMAND ${CMAKE_BINARY_DIR}/${TEST_NAME}
    DEPENDS ${TEST_NAME})

set_tests_properties(${TEST_NAME} PROPERTIES
    ENVIRONMENT "LD_LIBRARY_PATH=$ENV{LD_LIBRARY_PATH}:${CMAKE_BINARY_DIR}")

