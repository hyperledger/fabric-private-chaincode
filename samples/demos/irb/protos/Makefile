# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

PROTOS=$(basename $(wildcard *.proto))
GO_PROTO_BUILD_DIR := ../pkg/protos

all: build

$(addsuffix _pb2.py,$(PROTOS)):
	protoc -I=. --python_out=. $(patsubst %_pb2.py,%.proto,$@)

$(addsuffix .pb.go,$(PROTOS)): $(GO_PROTO_BUILD_DIR)
	protoc -I=. --go_out=$(GO_PROTO_BUILD_DIR) --go_opt=paths=source_relative $(patsubst %.pb.go,%.proto,$@)

build: $(addsuffix _pb2.py,$(PROTOS)) $(addsuffix .pb.go,$(PROTOS))

$(GO_PROTO_BUILD_DIR):
	mkdir -p $@

clean:
	rm -rf *.py $(GO_PROTO_BUILD_DIR)
