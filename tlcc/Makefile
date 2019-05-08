TOP = ..
include $(TOP)/build.mk

ENV_VARS := LD_LIBRARY_PATH=$(LD_LIBRARY_PATH):./enclave/lib
SRC := tlcc.go
OUTPUT := tlcc.so

$(OUTPUT): $(SRC)
	$(GO) get github.com/spf13/viper
	$(ENV_VARS) $(GO) build -o $@ -buildmode=plugin $<

build: $(OUTPUT)

test: build
	$(ENV_VARS) $(GO) $@ -test.v

clean:
	$(GO) clean
	rm -rf ledgersData
