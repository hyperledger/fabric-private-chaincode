#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

FROM node:8.16

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

COPY package.json /usr/src/app/package.json
COPY package-lock.json /usr/src/app/package-lock.json
ENV PATH /usr/src/app/node_modules/.bin:$PATH
WORKDIR /usr/src/app

# RUN npm rebuild

COPY gateway /usr/src/app/gateway
COPY src /usr/src/app/src
COPY config.json /usr/src/app/config.json
RUN npm install
CMD npm start
