.PHONY: all

PEER_NAME?=dev-jdoe
CC_NAME?=ecc
LD_LIB_PATH=$(LD_LIBRARY_PATH):./enclave/lib

all: build vscc-plugin

build:
	go get golang.org/x/sync/semaphore
	go get github.com/pkg/errors
	go get github.com/golang/protobuf/proto
	LD_LIBRARY_PATH=$(LD_LIB_PATH) go build

vscc-plugin:
	go build -o ./ecc-vscc.so -buildmode=plugin vscc/ecc_validation_plugin.go vscc/ecc_validation_logic.go

test:
	LD_LIBRARY_PATH=$(LD_LIB_PATH) go test -v

stress:
	LD_LIBRARY_PATH=$(LD_LIB_PATH) go test -v -run TestEnclaveChaincode_Invoke_Auction

debug: 
	LD_LIBRARY_PATH=$(LD_LIB_PATH) go test -c
	LD_LIBRARY_PATH=$(LD_LIB_PATH) sgx-gdb ecc.test -d $GOROOT

clean:
	go clean
	rm enclave/include/*
	rm enclave/lib/*
	rm enclave/mrenclave.go


DOCKER_CONTAINER=$(PEER_NAME)-$(CC_NAME)-0
DOCKER_IMAGE?=$$(docker images | grep -- -$(CC_NAME)- | awk '{print $$1;}')

docker:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run \
		-it \
		--device /dev/isgx \
		-v /var/run/aesmd:/var/run/aesmd \
		--name $(DOCKER_CONTAINER) \
		-e "CORE_CHAINCODE_LOGGING_LEVEL=DEBUG" \
		-e "CORE_CHAINCODE_LOGGING_SHIM=INFO" \
		-e "CORE_PEER_ADDRESS=172.17.0.1:7051" \
		-e "CORE_CHAINCODE_ID_NAME=ecc:0" \
		$(DOCKER_IMAGE) \
		chaincode

docker-clean:
	docker stop $(DOCKER_CONTAINER)
	docker rm $(DOCKER_CONTAINER)
