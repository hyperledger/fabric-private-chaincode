.PHONY: all

all: build vscc-plugin decorator-plugin

build:
	go build

vscc-plugin:
	go build -o ./ercc-vscc.so -buildmode=plugin vscc/ercc_validation_plugin.go vscc/ercc_validation_logic.go

decorator-plugin:
	go build -o ./ercc-decorator.so -buildmode=plugin attestation/ias_credentials/decoration.go

test:
	go test -v

clean: docker-clean
	go clean

ERCC_NAME?=ercc
DOCKER_IMAGE?=$$(docker images | grep -- -$(ERCC_NAME)- | awk '{print $$1;}')

docker-clean: docker-stop
	-docker rmi $(DOCKER_IMAGE)
