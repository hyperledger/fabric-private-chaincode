# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FROM node:12.13.1-alpine

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

ENV PATH /usr/src/app/node_modules/.bin:$PATH

COPY package.json /usr/src/app/
COPY package-lock.json /usr/src/app/
COPY src/ /usr/src/app/src
COPY .env.development /usr/src/app/

RUN npm install
RUN npm install -g @vue/cli

CMD ["npm", "run", "serve"]
