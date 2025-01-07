# FPC Playground for non-SGX environments

Here we explain how you can tinker with FPC even without Intel SGX hardware. That is, you can write your first FPC Chaincode compile it, and run it on almost any machine.

In your `config.override.mk` set the following to variables:
```Makefile
FPC_CCENV_IMAGE=ubuntu:22.04
ERCC_GOTAGS=
```
This configuration sets a standard Ubuntu image as alternative to our `fabric-private-chaincode-ccenv` image and overrides the default build tags we use to build `ercc`.

Next you can build `ercc` using the following command:
```bash
GOOS=linux make -C $FPC_PATH/ercc build docker
```

For building a chaincode, for instance `$FPC_PATH/samples/chaincode/kv-test-go`, just run: 
```bash
GOOS=linux make -C $FPC_PATH/samples/chaincode/kv-test-go with_go docker
```

You can test your FPC chaincode easily with one of the [sample deployments](../samples/deployment) tutorials.
We recommend to start with [the-simple-testing-network](../samples/deployment/fabric-smart-client/the-simple-testing-network).

Notes:
- On Mac use a recent version of bash (`brew install bash`).
- TODO more to come
