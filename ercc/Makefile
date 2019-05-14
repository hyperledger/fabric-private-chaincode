# Copyright 2019 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

TOP = ..
include $(TOP)/build.mk

VSCC_SRC = vscc/ercc_validation_plugin.go vscc/ercc_validation_logic.go
VSCC_OUT = ercc-vscc.so
DECO_SRC = attestation/ias_credentials/decoration.go
DECO_OUT = ercc-decorator.so

SUBDIRS = attestation

.PHONY: $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $(SUBDIRS)

$(SUBDIRS)-clean:
	$(MAKE) -C $(SUBDIRS) clean

build: $(SUBDIRS) ercc vscc-plugin decorator-plugin

ercc:
	$(GO) build

vscc-plugin: $(VSCC_SRC)
	$(GO) build -o $(VSCC_OUT) -buildmode=plugin $^

decorator-plugin: $(DECO_SRC)
	$(GO) build -o $(DECO_OUT) -buildmode=plugin $^

test: build
	$(GO) test -v

clean: docker-clean $(SUBDIRS)-clean
	$(GO) $@

ERCC_NAME?=ercc
DOCKER_IMAGE?=$$(docker images | grep -- -$(ERCC_NAME)- | awk '{print $$1;}')

docker-clean:
	-docker rmi $(DOCKER_IMAGE)
