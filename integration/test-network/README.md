# Setup the FPC test network

The FPC test network builds on the test-network provided by [fabric-samples](https://github.com/hyperledger/fabric-samples).

We provide fabric-samples as a submodule in `$FPC_PATH/integration/test-network/fabric-samples`.

Make sure you have installed [yq](https://github.com/mikefarah/yq).

Before you start the network make sure you build ercc and ecc_mock. In

```bash
cd $FPC_PATH/ercc
make && make docker

cd $FPC_PATH/ecc_mock
make && make docker
```

## Prepare network

Setup fabric sample network, binaries and docker images (this follow the [instructions](https://hyperledger-fabric.readthedocs.io/en/latest/install.html)).

```bash
cd $FPC_PATH/integration/test-network/fabric-samples
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.2.0 1.4.9 -s
```
 
```bash
cd $FPC_PATH/integration/test-network
./setup.sh
```

Start network:
```bash
cd $FPC_PATH/integration/test-network/fabric-samples/test-network
./network.sh up createChannel -c mychannel -ca -cai 1.4.9 -i 2.2.0
```

Install FPC components:
```bash
cd $FPC_PATH/integration/test-network
./installFPC.sh
# copy/past the export statement a successfully install will echo
```

Start FPC container
```bash
cd $FPC_PATH/integration/test-network
docker-compose up -d
```

Run test program
```bash
cd ${FPC_PATH}/client_sdk/go/test
make
./test
```

Shutdown network
```bash
cd $FPC_PATH/integration/test-network
docker-compose down
cd $FPC_PATH/integration/test-network/fabric-samples
./network.sh down
rm -f ${FPC_PATH}/client_sdk/go/test/wallet/appUser.id
```
