# FPC Playground for non-SGX environments

Here we explain how you can tinker with FPC even without Intel SGX hardware. That is, you can write your first FPC Chaincode, compile it, and run it on almost any machine, including a Mac with Apple Silicon. You need Go (the version in `go.mod`) and Docker.

Without SGX, FPC-go runs as plain Go and gives no confidentiality. Use it for development only.

This is the default: without `FPC_CGO=1`, the build uses plain Go instead of ego and a standard `ubuntu:22.04` base image instead of the SGX-based `fabric-private-chaincode-ccenv`, and ercc builds without the PDO crypto.

Build `ercc` and its image:
```bash
make -C $FPC_PATH/ercc docker
```

For building a chaincode, for instance `$FPC_PATH/samples/chaincode/kv-test-go`, run:
```bash
make -C $FPC_PATH/samples/chaincode/kv-test-go
make -C $FPC_PATH/samples/chaincode/kv-test-go docker
```
The first command compiles the chaincode for your machine. `docker` compiles it again for Linux and packages it, so the image also works when you build on a Mac.

You can test your FPC chaincode easily with one of the [sample deployments](../samples/deployment) tutorials.
We recommend to start with [the-simple-testing-network](../samples/deployment/fabric-smart-client/the-simple-testing-network).

Notes:
- On Mac use a recent version of bash (`brew install bash`).
