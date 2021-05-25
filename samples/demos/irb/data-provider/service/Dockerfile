# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FROM hyperledger/fabric-private-chaincode-ccenv:latest

RUN apt-get update && apt-get install -y poppler-utils

COPY data-provider-service /usr/local/bin

EXPOSE 3000

CMD ["data-provider-service"]

