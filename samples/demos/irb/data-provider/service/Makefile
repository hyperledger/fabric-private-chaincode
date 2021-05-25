# SPDX-License-Identifier: Apache-2.0

TOP = ../../../../..
include $(TOP)/build.mk

DOCKER_GOPATH=/project
DOCKER_FPC_PATH=$(DOCKER_GOPATH)/src/github.com/hyperledger/fabric-private-chaincode

build:
	$(GO) build $(GOTAGS) -o data-provider-service

docker:
	$(DOCKER) build $(DOCKER_BUILD_OPTS) \
    	--no-cache \
    	-t irb/data-provider-service:latest \
    	-f Dockerfile \
    	.

clean:
	rm -rf users
	rm -f data-provider-service

.PHONY: run
run:
	docker run --name data-provider-service \
				-v $(DOCKERD_FPC_PATH):$(DOCKER_FPC_PATH) \
				--env FPC_PATH=$(DOCKER_FPC_PATH) \
				--network host \
				-d irb/data-provider-service

.PHONY: stop
stop:
	docker stop data-provider-service
	docker rm data-provider-service