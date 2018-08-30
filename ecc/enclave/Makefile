.PHONY: all

all: build test

build:
	go build

test:
	LD_LIBRARY_PATH=./lib go test -test.v

clean:
	go clean