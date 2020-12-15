# Setup the FPC test network

The FPC test network builds on the test-network provided by [fabric-samples](https://github.com/hyperledger/fabric-samples).

We provide fabric-samples as a submodule in `$FPC_PATH/integration/test-network/fabric-samples`.

Make sure you have installed [yq](https://github.com/mikefarah/yq).
Note that you will version v3.4.1 or larger. 
For Ubuntu, `sudo snap install yq` is the easiest way to get a good version.

Before you start the network make sure you build ercc and ecc containers:

```bash
make build
```
If you want to build with mock-ecc rather than the real enclave-based one, build with
`make build GOTAGS="-tags mock_ecc"` instead.

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

## Do a test run
Start network:
```bash
cd $FPC_PATH/integration/test-network/fabric-samples/test-network
./network.sh up createChannel -c mychannel -ca -cai 1.4.9 -i 2.2.0
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
make ercc-ecc-run
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
