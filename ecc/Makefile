TOP = ..
include $(TOP)/build.mk

LD_LIBRARY_PATH := LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:./enclave/lib
PEER_NAME ?= dev-jdoe
CC_NAME ?= ecc

VSCC_SRC = vscc/ecc_validation_plugin.go vscc/ecc_validation_logic.go
VSCC_OUT = ecc-vscc.so

DOCKER_CONTAINER = $(PEER_NAME)-$(CC_NAME)-0
DOCKER_IMAGE ?= $$(docker images | grep -- -$(CC_NAME)- | awk '{print $$1;}')

#it seems that tests for ecc are broken;
#TODO fix tests and include test target to all
all: build
build: ecc vscc-plugin

ECC_ENCLAVE_BUILD = ../ecc_enclave/_build
ERR_MSG = "ecc_enclave build does not exist!"

ecc_enclave:
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/lib ]; then echo $(ERR_MSG); exit 1; fi
	@if [ ! -d $(ECC_ENCLAVE_BUILD)/include]; then echo $(ERR_MSG); exit 1; fi

ecc: ecc_enclave
	$(GO) get golang.org/x/sync/semaphore
	$(GO) get github.com/pkg/errors
	$(GO) get github.com/golang/protobuf/proto
	$(LD_LIBRARY_PATH) $(GO) build

vscc-plugin: $(VSCC_SRC)
	$(LD_LIBRARY_PATH) $(GO) build -o $(VSCC_OUT) -buildmode=plugin $^

# TODO fix test and enable this target
#test: ecc
#	$(LD_LIBRARY_PATH) $(GO) $@ -v

.PHONY: stress
stress:
	$(LD_LIBRARY_PATH) $(GO) test -v -run TestEnclaveChaincode_Invoke_Auction

.PHONY: debug
debug:
	$(LD_LIBRARY_PATH) $(GO) test -c
	$(LD_LIBRARY_PATH) sgx-gdb ecc.test -d $(GOROOT)

clean: docker-clean
	$(GO) clean
#	rm enclave/include/*
#	rm enclave/lib/*
#	rm enclave/mrenclave.go

docker:
	$(DOCKER) build -t $(DOCKER_IMAGE) -f Dockerfile ..

docker-run:
	$(DOCKER) run \
		-it \
		--device /dev/isgx \
		-v /var/run/aesmd:/var/run/aesmd \
		--name $(DOCKER_CONTAINER) \
		-e "CORE_CHAINCODE_LOGGING_LEVEL=DEBUG" \
		-e "CORE_CHAINCODE_LOGGING_SHIM=INFO" \
		-e "CORE_PEER_ADDRESS=172.17.0.1:7051" \
		-e "CORE_CHAINCODE_ID_NAME=$(CC_NAME):0" \
		$(DOCKER_IMAGE) \
		chaincode

docker-stop:
	-docker stop $(DOCKER_CONTAINER)
	-docker rm $(DOCKER_CONTAINER)

docker-clean: docker-stop
	-docker rmi $(DOCKER_IMAGE)

