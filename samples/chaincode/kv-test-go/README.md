# KV test (FPC-go)

A key-value store as an FPC chaincode written in Go, used by the integration tests. It is the Go counterpart of the C++ [kv-test](../kv-test) sample.

## Run without SGX

With only Go and Docker (also on a Mac):
```bash
make             # compile with plain go
make docker      # build the fpc/kv-test-go image
```
Then deploy it on the Fabric test network as described in [samples/deployment/test-network](../../deployment/test-network/README.md#without-sgx-eg-on-a-mac), with `CC_ID=kv-test-go` and `CC_NAME=kv-test-go`, or run it with the [the-simple-testing-network](../../deployment/fabric-smart-client/the-simple-testing-network) sample.

Without SGX, FPC gives no confidentiality; the chaincode says so in its log at startup. Use this for development only. FPC-go's attestation is always simulated today.

## Run with SGX

In the [dev container](../../../docs/setup-option1.md), `make` builds the chaincode with ego and packages it into an SGX image.
