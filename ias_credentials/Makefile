DECO_SRC = attestation/ias_credentials/decoration.go
DECO_OUT = ercc-decorator.so

decorator-plugin: $(DECO_SRC)
	$(GO) build -o $(DECO_OUT) -buildmode=plugin $^
