# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ..
include $(TOP)/build.mk

PLANTUML_SRCS = fpc-lifecycle.txt
PLANTUML_IMGS = $(PLANTUML_SRCS:.txt=.$(PLANTUML_IMG_FORMAT))

all: build

build : $(PLANTUML_IMGS)

clean :
	-rm $(PLANTUML_IMGS)

%.$(PLANTUML_IMG_FORMAT): %.txt
	$(PLANTUML_CMD) -t$(PLANTUML_IMG_FORMAT) $<


