# FPC and the Fabric-Samples Test Network

This guide shows how to deploy and run a FPC Chaincode on test-network provided by [fabric-samples](https://github.com/hyperledger/fabric-samples).
We provide fabric-samples as a submodule in `$FPC_PATH/samples/deployment/test-network/fabric-samples`.

Before moving forward, follow the main [README](../../../README.md) to set up your environment. This guide works also with
the FPC Dev docker container.

## Prepare FPC Containers and the Test Network

FPC requires a special docker container to execute a FPC chaincode, similar to Fabric's `ccenv` container image but with additional support for Intel SGX.  
You can pull the FPC chaincode environment image (`fabric-private-chaincode-ccenv`) from our Github repository or build them manually as follows:

```bash
# pulls the fabric-private-chaincode-ccenv image from github
make -C $FPC_PATH/utils/docker pull

# building fabric-private-chaincode-ccenv image from scratch
make -C $FPC_PATH/utils/docker build
```

Next, we package the FPC components as docker images (building on top of `fabric-private-chaincode-ccenv`) which are deployed on our test network.
Use `CC_ID` and `CC_PATH` to define the FPC Chaincode you want to build.

```bash
cd $FPC_PATH/samples/deployment/test-network
export CC_ID=echo
export CC_PATH=$FPC_PATH/samples/chaincode/echo
make build
```

Note: If you want to build with [mock-enclave](../../../ecc/chaincode/enclave/mock_enclave.go) rather than the real enclave-based one, build with
`make build GOTAGS="-tags mock_ecc"` instead.

Next, setup fabric sample network, binaries and docker images. Here we follow the official Fabric [instructions](https://hyperledger-fabric.readthedocs.io/en/release-2.3/install.html).

```bash
cd $FPC_PATH/samples/deployment/test-network
git clone https://github.com/hyperledger/fabric-samples
cd $FPC_PATH/samples/deployment/test-network/fabric-samples
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.3.3 1.4.9 -s
```

Before we can start the network, we need to update the Fabric peer configuration to enable FPC support.
That is, we need to install the external builders by adding the following lines to the peers `core.yaml`:
```yaml
chaincode:
  externalBuilders:
    - path: /opt/gopath/src/github.com/hyperledger/fabric-private-chaincode/fabric/externalBuilder/chaincode_server
      name: fpc-external-launcher
      propagateEnvironment:
        - CORE_PEER_ID
        - FABRIC_LOGGING_SPEC
```

Since the Fabric-Samples test network uses docker and docker-compose, we need to ensure that the external Builder
scripts are also available inside the peer container. For this reason, we mount `$FPC_PATH/fabric/externalBuilder/chaincode_server` into
`/opt/gopath/src/github.com/hyperledger/fabric-private-chaincode/fabric/externalBuilder/chaincode_server`.

For convenience, we provide a `setup.sh` script to update the `core.yaml` and the docker compose files to mount the external
Builder.

```bash
cd $FPC_PATH/samples/deployment/test-network
./setup.sh
```

## Start the Network

Let's start the Fabric-Samples test network.
```bash
cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
./network.sh up -ca
./network.sh createChannel -c mychannel
```

Next, we install the FPC Enclave Registry and our FPC Chaincode on the network by using the standard Lifecycle commands,
including chaincode `package`, `install`, `approveformyorg`, and `commit`. An important detail here is that the
chaincode packaging differs from traditional chaincode. Since we are using the external Builder and run the FPC Chaincode
as `Chaincode as a Service (CaaS)`, the packaging artifact will contain information including the CaaS endpoint rather
than the actual Chaincode.
We continue with the following command to install the FPC Chaincode as just described.
```bash
cd $FPC_PATH/samples/deployment/test-network
./installFPC.sh
# IMPORTANT: a successfully install will show you an `export ...`
# statement as stdout on the command-line.  Copy/Paste this statement
# into your shell or below starting of FPC containers will not work properly
# (but also would not give you clear errors that it doesn't!!)
```

Now we have the FPC Chaincode installed on the Fabric peers, but we still need to start our chaincode containers.
Make sure you have set `CC_ID` to the same chaincode ID as used in the earlier step when building the chaincode. Also confirm that `CC_PATH` is set to the location of the FPC chaincode code.

```bash
# Start FPC container
make ercc-ecc-start
```
You should see two instances of the FPC Echo chaincode and two instances of the FPC Enclave Registry chaincode running such as in the following example:
```bash
Creating ercc.peer0.org2.example.com ... done
Creating ercc.peer0.org1.example.com ... done
Creating echo.peer0.org2.example.com ... done
Creating echo.peer0.org1.example.com ... done
Attaching to echo.peer0.org2.example.com, echo.peer0.org1.example.com, ercc.peer0.org1.example.com, ercc.peer0.org2.example.com
echo.peer0.org1.example.com    | 2021-05-11 16:11:31.342 UTC [ecc] main -> INFO 001 starting fpc chaincode (echo_16B510D1D2581EA4679A2D0D1C50A7C3BE87282324D5AB4DBAA89C1CC4832C85:73e7a0e342f5ea85ec2809a4c9423f17ad0388824be4e85ce02d0b80b6d39723)
echo.peer0.org2.example.com    | 2021-05-11 16:11:31.112 UTC [ecc] main -> INFO 001 starting fpc chaincode (echo_16B510D1D2581EA4679A2D0D1C50A7C3BE87282324D5AB4DBAA89C1CC4832C85:fcc2e8287c940009011a87830a1f5ca3085c4412fb86fa86060a094cac1e5f03)
ercc.peer0.org2.example.com    | 2021-05-11 16:11:32.107 UTC [cgo] 0 -> INFO 001 Initializing logger
ercc.peer0.org1.example.com    | 2021-05-11 16:11:32.108 UTC [cgo] 0 -> INFO 001 Initializing logger
ercc.peer0.org2.example.com    | 2021-05-11 16:11:32.170 UTC [ercc] main -> INFO 002 starting enclave registry (ercc_1.0:db2e97768a87d2b9ed9d86a729a767ef1040fb2e20d2f2a907e727ece7256028)
ercc.peer0.org1.example.com    | 2021-05-11 16:11:32.199 UTC [ercc] main -> INFO 002 starting enclave registry (ercc_1.0:0919bd1be2e779a582a537da8e29fba241972c3531472006b170b0c26a1d71fb)
```
The FPC Chaincode is now up and running, ready for processing invocations!
Note that the containers are running in foreground in your terminal using docker-compose.

## Interact with the FPC Chaincode

Now we show how to use the [FPC Client SDK](../../../client_sdk/go) to interact with the FPC Chaincode running on the test network.
We continue with a new terminal to keep the FPC Chaincode running in the other terminal as mentioned before.

The Fabric-Samples test network generates the connection profiles which are required by the FPC Client SDK to connect to
the network. For example, you can find the connection profile for `org1` in
`$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/`.
However, the generated connection profiles are missing some additional information to be used with FPC, in particular, to use
the [LifecycleInitEnclave](../../../client_sdk/go/pkg/client/resmgmt/lifecycleclient.go) command.
Moreover, FPC Client SDK currently requires the connection profile to contain the connection details of the peer that hosts the FPC Chaincode Enclave. We use a helper script to update the connect profile files.

```bash
cd $FPC_PATH/samples/deployment/test-network
./update-connection.sh
```

### How to use simple-go

Now we will use the go app in `$FPC_PATH/samples/application/simple-go` to demonstrate the usage of the FPC Client SDK.
In order to initiate the FPC Chaincode enclave and register it with the FPC Enclave Registry, run the app with the `-withLifecycleInitEnclave` flag.

```bash
cd $FPC_PATH/samples/application/simple-go
CC_ID=echo ORG_NAME=Org1 go run . -withLifecycleInitEnclave
```
Note that we execute the go app as `Org1`, thereby creating and registering the FPC Chaincode enclave at `peer0.org1.example.com`. Alternatively, we could run this as `Org2` to initiate the enclave at `peer0.org2.example.com`.

Afterwards you _must_ run the application without the `withLifecycleInitEnclave` flag and you can play with multiple organizations.
```bash
cd $FPC_PATH/samples/application/simple-go
CC_ID=echo ORG_NAME=Org1 go run .
CC_ID=echo ORG_NAME=Org2 go run .
```

### How to use simple-cli-go

You can also use `$FPC_PATH/samples/application/simple-cli-go` instead of the simple-go application.

```bash
# make fpcclient
cd $FPC_PATH/samples/application/simple-cli-go
make

# export fpcclient settings
export CC_NAME=echo
export CHANNEL_NAME=mychannel
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_ID=peer0.org1.example.com
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_MSPCONFIGPATH=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_TLS_CERT_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt
export CORE_PEER_TLS_ENABLED="true"
export CORE_PEER_TLS_KEY_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
export CORE_PEER_TLS_ROOTCERT_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export ORDERER_CA=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export GATEWAY_CONFIG=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml

# init our enclave
./fpcclient init $CORE_PEER_ID

# interact with the FPC Chaincode
./fpcclient invoke foo
./fpcclient query foo
```

## Shutdown network
Since we opened a new terminal to interact with the FPC Chaincode, to be able to shutdown the FPC chaincode you need to define the environment variables that set the chaincode name and path.
```bash
export CC_ID=echo
export CC_PATH=$FPC_PATH/samples/chaincode/echo
make -C $FPC_PATH/samples/deployment/test-network ercc-ecc-stop
cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
./network.sh down
```

## Using HelloWorld chaincode on the test-network
In this section we show how you can run a different FPC chaincode on the test network. For example, you can use the [HelloWorld](../../chaincode/helloworld/README.md) code instead of the echo code. To do so you have to change the values of the environment variables set at the beginning `CC_ID` and `CC_PATH`.
```bash
export CC_ID=helloworld
export CC_PATH=$FPC_PATH/samples/chaincode/helloworld
```
Afterwards to test you would use [simple-cli-go](#How-to-use-simple-cli-go). You would need to verify that the `CC_NAME` variable is set to helloworld and then to interact you would execute the chaincode as follows:
```bash
# interact with the FPC Chaincode
./fpcclient invoke storeAsset asset1 100
./fpcclient query retrieveAsset asset1
```

## Debugging

For diagnostics, you can run the following to see logs for `peer0.org1.example.com`.
```bash
docker logs -f peer0.org1.example.com
docker logs -f ercc.peer0.org1.example.com
docker logs -f ecc.peer0.org1.example.com
```

To interact with the peer using the `peer CLI`, run the following
```bash
cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network;
export FABRIC_CFG_PATH=$FPC_PATH/samples/deployment/test-network/fabric-samples/config
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

## Using Blockchain Explorer

Another way to illustrate the use of FPC is using [Hyperledger Blockchain Explorer](https://github.com/hyperledger/blockchain-explorer) with the test network.
This tool allows you to see the transactions committed to the ledger.
When you inspect FPC transactions, you will notice that the content of the read/writeset is encrypted.

We have already integrated Blockchain Explorer in the test network using our `setup.sh` script.
You can find the Blockchain Explorer configuration files in the `blockchain-explorer` folder after running `setup.sh`.
Note that `setup.sh` may ask you to override any existing configuration files in `blockchain-explorer` and restore the default configuration.  

To start Blockchain Explorer we use docker compose. Just run the following
```bash
cd $FPC_PATH/samples/deployment/test-network/blockchain-explorer
docker-compose up -d
```

Once it is up and running you can access the web interface using your browser.
The url is `http://localhost:8080/`. To log in, use the username `exploreradmin` and the password `exploreradminpw`.
