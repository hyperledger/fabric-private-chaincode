.PHONY: all

all: build vscc-plugin

build:
	go build

vscc-plugin:
	go build -o ./ercc-vscc.so -buildmode=plugin vscc/ercc_validation_plugin.go vscc/ercc_validation_logic.go

test:
	go test -v

clean:
	go clean

