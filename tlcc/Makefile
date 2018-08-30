.PHONY: all

PEER_NAME?=dev-jdoe
LD_LIB_PATH=$(LD_LIBRARY_PATH):./enclave/lib


all: build build-plugin test

build:
	LD_LIBRARY_PATH=$(LD_LIB_PATH) go build

build-plugin:
	LD_LIBRARY_PATH=$(LD_LIB_PATH) go build -o ./tlcc.so -buildmode=plugin tlcc.go

test: build-plugin
	LD_LIBRARY_PATH=$(LD_LIB_PATH) go test -test.v

run: build
	LD_LIBRARY_PATH=$(LD_LIB_PATH) CORE_CHAINCODE_LOGGING_LEVEL=DEBUG CORE_CHAINCODE_LOGGING_SHIM=INFO CORE_PEER_ADDRESS=localhost:7051 CORE_CHAINCODE_ID_NAME=ecc:0 ./tlcc

clean:
	go clean
	rm enclave/include/*
	rm enclave/lib/*
	rm enclave/mrenclave.go
	rm -rf ledgersData
