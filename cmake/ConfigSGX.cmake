# get arch information
execute_process(COMMAND getconf LONG_BIT OUTPUT_VARIABLE VAR_LONG_BIT)

# Set variables for common SGX variables.
# if they are defined in the environment, take them from there, otherwise
# set them to the usual defaults.
set(SGX_SDK "$ENV{SGX_SDK}")
if("${SGX_SDK} " STREQUAL " ")
       set(SGX_SDK /opt/intel/sgxsdk)
endif()

set(SGX_ARCH "$ENV{SGX_ARCH}")
if("${SGX_ARCH} " STREQUAL " ")
       set(SGX_ARCH x64)
endif()

set(SGX_MODE "$ENV{SGX_MODE}")
if("${SGX_MODE} " STREQUAL " ")
       set(SGX_MODE HW) # SGX mode: sim, hw
endif()

set(SGX_BUILD "$ENV{SGX_BUILD}")
if("${SGX_BUILD} " STREQUAL " ")
       set(SGX_BUILD PRERELEASE)
endif()

set(SGX_SSL "$ENV{SGX_SSL}")
if("${SGX_SSL} " STREQUAL " ")
       set(SGX_SSL /opt/intel/sgxssl)
endif()

set(SGX_COMMON_CFLAGS -m64)
set(SGX_LIBRARY_PATH ${SGX_SDK}/lib64)
set(SGX_ENCLAVE_SIGNER ${SGX_SDK}/bin/x64/sgx_sign)
set(SGX_EDGER8R ${SGX_SDK}/bin/x64/sgx_edger8r)

if (SGX_MODE STREQUAL HW)
    set(SGX_URTS_LIB sgx_urts)
    set(SGX_USVC_LIB sgx_uae_service)
    set(SGX_TRTS_LIB sgx_trts)
    set(SGX_TSVC_LIB sgx_tservice)
else ()
    set(SGX_URTS_LIB sgx_urts_sim)
    set(SGX_USVC_LIB sgx_uae_service_sim)
    set(SGX_TRTS_LIB sgx_trts_sim)
    set(SGX_TSVC_LIB sgx_tservice_sim)
endif (SGX_MODE STREQUAL HW)

if (SGX_BUILD STREQUAL "DEBUG")
    set(SGX_COMMON_CFLAGS "${SGX_COMMON_CFLAGS} -O0 -g")
    set(SGX_COMMON_CFLAGS "${SGX_COMMON_CFLAGS} -DDO_DEBUG=true")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -DDEBUG -UNDEBUG -UEDEBUG")
elseif(SGX_BUILD STREQUAL "PRERELEASE")
    set(SGX_COMMON_CFLAGS "${SGX_COMMON_CFLAGS} -O2")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -UDEBUG -DNDEBUG -DEDEBUG")
elseif(SGX_BUILD STREQUAL "RELEASE")
    if(SGX_MODE STREQUAL "HW")
        set(SGX_COMMON_CFLAGS "${SGX_COMMON_CFLAGS} -O2")
        set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -UDEBUG -DNDEBUG -UEDEBUG")
    else()
        message(FATAL_ERROR "HW mode must be set with RELEASE")
    endif()
else()
    message(FATAL_ERROR "Unknown build ${SGX_BUILD}")
endif()

set(SGX_SSL_LIBRARY_PATH ${SGX_SSL}/lib64)

message(STATUS "SGX_COMMON_CFLAGS: ${SGX_COMMON_CFLAGS}")
message(STATUS "SGX_SDK: ${SGX_SDK}")
message(STATUS "SGX_MODE: ${SGX_MODE}")
message(STATUS "SGX_BUILD: ${SGX_BUILD}")
message(STATUS "SGX_LIBRARY_PATH: ${SGX_LIBRARY_PATH}")
message(STATUS "SGX_ENCLAVE_SIGNER: ${SGX_ENCLAVE_SIGNER}")
message(STATUS "SGX_EDGER8R: ${SGX_EDGER8R}")
message(STATUS "SGX_SSL: ${SGX_SSL}")
message(STATUS "SGX_SSL_LIBRARY_PATH: ${SGX_SSL_LIBRARY_PATH}")
