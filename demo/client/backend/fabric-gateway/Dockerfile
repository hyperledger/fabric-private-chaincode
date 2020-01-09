#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

FROM node:8.16

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

ENV PATH /usr/src/app/node_modules/.bin:$PATH

COPY \
    package.json \
    package-lock.json\
    config.json\
    auction.json\
    fabric-sdk-node_fabric-network_lib_transaction.js.patch\
  /usr/src/app/
COPY src /usr/src/app/src
COPY gateway /usr/src/app/gateway

RUN npm install \
  &&  patch -R -p0 < ./fabric-sdk-node_fabric-network_lib_transaction.js.patch

CMD npm start
