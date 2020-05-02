# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../../..
include $(TOP)/build.mk

PLANTUML_SRCS = $(wildcard *.puml)
PLANTUML_IMGS = $(PLANTUML_SRCS:.puml=.$(PLANTUML_IMG_FORMAT))

all: build

build : $(PLANTUML_IMGS)

clean :
	rm $(PLANTUML_IMGS)

%.$(PLANTUML_IMG_FORMAT): %.puml
	PLANTUML_LIMIT_SIZE=8192 $(PLANTUML_CMD) -t$(PLANTUML_IMG_FORMAT) $<


