# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

DOCKER_IMAGE = auction_frontend
APP_SRC = ${PWD}

build:
	$(DOCKER) build -t $(DOCKER_IMAGE) .

run:
	$(DOCKER) run \
		-it \
		-v ${APP_SRC}:/usr/src/app \
		-v /usr/src/app/node_modules \
		-p 5000:5000 \
		$(DOCKER_IMAGE)

cli:
	$(DOCKER) run \
		-it \
		-v ${APP_SRC}:/usr/src/app \
		$(DOCKER_IMAGE) \
		/bin/sh
