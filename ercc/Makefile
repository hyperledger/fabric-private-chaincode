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

clean:
	go clean

