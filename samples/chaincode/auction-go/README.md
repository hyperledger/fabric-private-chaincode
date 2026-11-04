# Auction (FPC-go)

A sealed-bid auction as an FPC chaincode written in Go, with the functions `init`, `create`, `submit`, `close` and `eval`. It is the Go counterpart of the C++ [auction](../../../sgx/samples/chaincode/auction) sample.

## Run without SGX

With only Go and Docker (also on a Mac):
```bash
make             # compile with plain go
make docker      # build the fpc/auction-go image
```
Then deploy it on the Fabric test network as described in [samples/deployment/test-network](../../deployment/test-network/README.md#without-sgx-eg-on-a-mac), with `CC_ID=auction-go` and `CC_NAME=auction-go`.

Without SGX, FPC gives no confidentiality; the chaincode says so in its log at startup. Use this for development only. FPC-go's attestation is always simulated today.

## Run with SGX

In the [dev container](../../../docs/setup-option1.md), `make` builds the chaincode with ego and packages it into an SGX image.
