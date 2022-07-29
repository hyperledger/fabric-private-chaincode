# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Description:
#   Sets up the template of a docker environment to run FPC go chaincode
#
#  Configuration (build) parameters (for defaults, see below section with ARGs)
#  - fpc image version:          FPC_VERSION
#  - sgxmode:                    SGX_MODE
#  - Chaincode as a Server Port: CAAS_PORT

ARG FPC_CCENV_IMAGE=hyperledger/fabric-private-chaincode-ccenv:main

FROM ${FPC_CCENV_IMAGE}

ARG SGX_MODE
ENV SGX_MODE=${SGX_MODE}
ARG OE_SIMULATION
ENV OE_SIMULATION=${OE_SIMULATION}

ARG CAAS_PORT=9999
ARG CC_PATH="/usr/local/bin"

COPY ecc-bundle ${CC_PATH}/chaincode

WORKDIR ${CC_PATH}
EXPOSE ${CAAS_PORT}

CMD ["chaincode"]
