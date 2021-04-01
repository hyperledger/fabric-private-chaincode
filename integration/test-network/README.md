# FPC and the Fabric-Samples Test Network

This guide shows how to deploy and run a FPC Chaincode on test-network provided by [fabric-samples](https://github.com/hyperledger/fabric-samples).
We provide fabric-samples as a submodule in `$FPC_PATH/integration/test-network/fabric-samples`.

Before moving forward, follow the main [README](../../README.md) to set up your environment. This guide works also with
the FPC Dev docker container.

## Prepare FPC Containers and the Test Network

We start with building the FPC components as docker images which are deployed on our test network.
Use `TEST_CC_ID` and `TEST_CC_PATH` to define the FPC Chaincode you want to build. 

```bash
cd $FPC_PATH/integration/test-network/
export TEST_CC_ID=echo
export TEST_CC_PATH=${FPC_PATH}/examples/echo
make build
```

Note: If you want to build with [mock-enclave](../../ecc/chaincode/enclave/mock_enclave.go) rather than the real enclave-based one, build with
`make build GOTAGS="-tags mock_ecc"` instead.

Next, setup fabric sample network, binaries and docker images. Here we follow the official Fabric [instructions](https://hyperledger-fabric.readthedocs.io/en/latest/install.html).

```bash
cd $FPC_PATH/integration/test-network/fabric-samples
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.3.0 1.4.9 -s
```

Before we can start the network, we need to update the Fabric peer configuration to enable FPC support.
That is, we need to install the external builders by adding the following lines to the peers `core.yaml`:
```yaml
chaincode:
  externalBuilders:
    - path: /opt/gopath/src/github.com/hyperledger-labs/fabric-private-chaincode/fabric/externalBuilder/chaincode_server
      name: fpc-external-launcher
      propagateEnvironment:
        - CORE_PEER_ID
        - FABRIC_LOGGING_SPEC
```

Since the Fabric-Samples test network uses docker and docker-compose, we need to ensure that the external Builder
scripts are also available inside the peer container. For this reason, we mount `$FPC_PATH/fabric/externalBuilder/chaincode_server` into
`/opt/gopath/src/github.com/hyperledger-labs/fabric-private-chaincode/fabric/externalBuilder/chaincode_server`.

For convenience, we provide a `setup.sh` script to update the `core.yaml` and the docker compose files to mount the external
Builder.

```bash
cd $FPC_PATH/integration/test-network
./setup.sh
```

## Start the Network

Let's start the Fabric-Samples test network.
```bash
cd $FPC_PATH/integration/test-network/fabric-samples/test-network
./network.sh up 
./network.sh createChannel -c mychannel -ca -cai 1.4.9 -i 2.3.0
```

Next, we install the FPC Enclave Registry and our FPC Chaincode on the network by using the standard Lifecycle commands,
including chaincode `package`, `install`, `approveformyorg`, and `commit`. An important detail here is that the
chaincode packaging differs from traditional chaincode. Since we are using the external Builder and run the FPC Chaincode
as `Chaincode as a Service (CaaS)`, the packaging artifact will contain information including the CaaS endpoint rather
than the actual Chaincode.
We continue with the following command to install the FPC Chaincode as just described.
```bash
cd $FPC_PATH/integration/test-network
./installFPC.sh
# IMPORTANT: a successfully install will show you an `export ...`
# statement as stdout on the command-line.  Copy/Paste this statement
# into your shell or below starting of FPC containers will not work properly
# (but also would not give you clear errors that it doesn't!!)
```

Now we have the FPC Chaincode installed on the Fabric peers, but we still need to start our chaincode containers. 
Make sure you have set `TEST_CC_ID` to the same chaincode ID as used in the earlier step when building the chaincode.

```bash
# Start FPC container
export TEST_CC_ID=echo
make ercc-ecc-start
```

The FPC Chaincode is now up and running, ready for processing invocations!

## Interact with the FPC Chaincode

Now we show how to use the [FPC Client SDK](../../client_sdk/go) to interact with the FPC Chaincode running on the test network.

The Fabric-Samples test network generates the connection profiles which are required by the FPC Client SDK to connect to
the network. For example, you can find the connection profile for org1 in
`$FPC_PATH/integration/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/`.
However, the generated connection profiles are supposed to be only used with the gateway API. For this reason,   
we need to prepare the `connection-*.yaml` files to support the FPC [LifecycleInitEnclave](../../client_sdk/go/pkg/client/resmgmt/lifecycleclient.go)
command as it builds on the Fabric SDK Go Low-level API. 

```bash
cd $FPC_PATH/integration/test-network
./update-connection.sh
```

Now we will use the go app in `${FPC_PATH}/client_sdk/go/sample` to demonstrate the usage of the FPC Client SDK.

In order to initiate the FPC Chaincode enclave and register it with the FPC Enclave Registry, run the app with the `-withLifecycleInitEnclave` flag.

```bash
cd ${FPC_PATH}/client_sdk/go/sample
CC_ID=echo go run . -withLifecycleInitEnclave
```

Afterwards you can run the application without the flag
```bash
cd ${FPC_PATH}/client_sdk/go/sample
CC_ID=echo go run .
```

## Shutdown network
```bash
make -C $FPC_PATH/integration/test-network ercc-ecc-stop
cd $FPC_PATH/integration/test-network/fabric-samples/test-network
./network.sh down
rm -f ${FPC_PATH}/client_sdk/go/sample/wallet/appUser.id
```

## Debugging

For diagnostics, you can run the following to see logs for peer0.org1 ...
```bash
docker logs -f peer0.org1.example.com
docker logs -f ercc.peer0.org1.example.com
docker logs -f ecc.peer0.org1.example.com
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
