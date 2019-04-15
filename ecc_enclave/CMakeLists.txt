cmake_minimum_required(VERSION 3.5.1)
project(ecclib)

include(../cmake/Init.cmake)

include(${CMAKE_MODULE_PATH}/ConfigSGX.cmake)

set(COMMON_SOURCE_DIR ${PROJECT_SOURCE_DIR}/../common)

add_subdirectory(enclave)
add_subdirectory(sgxcclib)

set(BUILD_DIR ${CMAKE_LIBRARY_OUTPUT_DIRECTORY}/include ${CMAKE_BINARY_DIR}/lib)
set(DEPLOY_DIR ${PROJECT_SOURCE_DIR}/../ecc/enclave)

add_custom_target(deploy
    COMMAND cp -rv ${BUILD_DIR} ${CMAKE_BINARY_DIR}/mrenclave.go ${DEPLOY_DIR}
    COMMENT "Deploy ecclib to ${DEPLOY_DIR}"
    DEPENDS enclave sgxcclib)
