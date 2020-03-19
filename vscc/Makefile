VSCC_SRC = vscc/ercc_validation_plugin.go vscc/ercc_validation_logic.go
VSCC_OUT = ercc-vscc.so

vscc-plugin: $(VSCC_SRC)
	$(GO) build -o $(VSCC_OUT) -buildmode=plugin $^
