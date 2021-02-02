# Setup the FPC test network

The FPC test network builds on the test-network provided by [fabric-samples](https://github.com/hyperledger/fabric-samples).
We provide fabric-samples as a submodule in `$FPC_PATH/integration/test-network/fabric-samples`.

In order to use FPC with the test network, make sure you have installed [yq](https://github.com/mikefarah/yq) (version 3.x).
Not that newer versions (v4.x and higher) are currently not supported.
You can install `yq` v3 via `go get`.
```bash
GO111MODULE=on go get github.com/mikefarah/yq/v3
```

In addition to `yq` you need a recent version of docker-compose (version 1.25 or higher).
Note Ubuntu 18.04 comes with an older version of docker-compose and thus needs to be updated.
See related notes in [Working from behind a proxy](../../README.md#working-from-behind-a-proxy) in our [README.md](../../README.md) for more information.

[comment]: <> (This comment can be removed with upgrading FPC to general Ubuntu 20.04 support)
If you run the FPC test network from within our FPC docker dev container, please use Ubuntu 20.04 Docker images.
```Makefile
DOCKER_BUILD_OPTS="--build-arg UBUNTU_VERSION=20.04 --build-arg UBUNTU_NAME=focal" make -C $FPC_PATH/utils/docker run
```

## Prepare FPC containers and network

Before you start the network make sure you build ercc and ecc containers:

```bash
cd $FPC_PATH/integration/test-network/
make build
```
If you want to build with mock-ecc rather than the real enclave-based one, build with
`make build GOTAGS="-tags mock_ecc"` instead.

Setup fabric sample network, binaries and docker images (this follow the [instructions](https://hyperledger-fabric.readthedocs.io/en/latest/install.html)).

```bash
cd $FPC_PATH/integration/test-network/fabric-samples
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.3.0 1.4.9 -s
```
 
```bash
cd $FPC_PATH/integration/test-network
./setup.sh
```

## Do a test run
Start network:
```bash
cd $FPC_PATH/integration/test-network/fabric-samples/test-network
./network.sh up createChannel -c mychannel -ca -cai 1.4.9 -i 2.3.0
```

Install FPC components:
```bash
cd $FPC_PATH/integration/test-network
./installFPC.sh
# IMPORTANT: a successfully install will show you an `export ...`
# statement as stdout on the command-line.  Copy/Paste this statement
# into your shell or below starting of FPC containers will not work properly
# (but also would not give you clear errors that it doesn't!!)
```

Start FPC container
```bash
make ercc-ecc-start
```

Run test program
```bash
cd ${FPC_PATH}/client_sdk/go/test
make
make run
```

Shutdown network
```bash
make -C $FPC_PATH/integration/test-network ercc-ecc-stop
cd $FPC_PATH/integration/test-network/fabric-samples/test-network
./network.sh down
rm -f ${FPC_PATH}/client_sdk/go/test/wallet/appUser.id
```

For diagnostics, you can run the following to see peer logs
```bash
docker logs -f peer0.org1.example.com
```

To interact interactively with the peer, run the following
```bash
cd $FPC_PATH/integration/test-network/fabric-samples/test-network;
export FABRIC_CFG_PATH=$FPC_PATH/integration/test-network/fabric-samples/config
export PATH=$(readlink -f ../bin):$PATH
source ./scripts/envVar.sh; \
setGlobals 1;
```
and you will be able to run the usual peer cli commands, e.g.,
```bash
peer lifecycle chaincode queryinstalled
```
and, in particular, also access ercc to see the registry state, e.g.,
```bash
peer chaincode query -C mychannel -n ercc -c '{"Function": "queryChaincodeEndPoints", "Args" : ["echo"]}'
peer chaincode query -C mychannel -n ercc -c '{"Function": "queryListProvisionedEnclaves", "Args" : ["echo"]}'
peer chaincode query -C mychannel -n ercc -c '{"Function": "queryChaincodeEncryptionKey", "Args" : ["echo"]}'
peer chaincode query -C mychannel -n ercc -c '{"Function": "queryListEnclaveCredentials", "Args" : ["echo"]}'
E_ID=$(peer chaincode query -C mychannel -n ercc -c '{"Function": "queryListProvisionedEnclaves", "Args" : ["echo"]}' 2> /dev/null  | jq -r '.[0]')
peer chaincode query -C mychannel -n ercc -c '{"Function": "queryEnclaveCredentials", "Args" : ["echo", "'${E_ID}'"]}'
```
