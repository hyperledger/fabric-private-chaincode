FROM hyperledger/fabric-ccenv-sgx:latest

ARG CC_NAME="ecc"
ARG CC_PATH="/usr/local/bin"
ARG CC_LIB_PATH=${CC_PATH}"/enclave/lib"

RUN mkdir -p ${CC_LIB_PATH}

ENV LD_LIBRARY_PATH=${CC_LIB_PATH}

copy ${CC_NAME} ${CC_PATH}/chaincode
copy enclave/lib/*.so ${CC_LIB_PATH}/

WORKDIR ${CC_PATH}
