# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

PROTOS=$(basename $(wildcard *.proto))

all: build

$(addsuffix _pb2.py,$(PROTOS)):
	protoc -I=. --python_out=. $(patsubst %_pb2.py,%.proto,$@)

$(addsuffix .pb.go,$(PROTOS)):
	protoc -I=. --go_out=${GOPATH}/src $(patsubst %.pb.go,%.proto,$@)

build: $(addsuffix _pb2.py,$(PROTOS)) $(addsuffix .pb.go,$(PROTOS))
#build: $(addsuffix .pb.go,$(PROTOS))

clean:
	rm -rf *.py *.go
