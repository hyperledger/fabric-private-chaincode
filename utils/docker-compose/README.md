# FPC Docker Compose Network
This docker-compose example has been adapted from a [Fabric 101 Workshop](https://github.com/swetharepakula/Fabric101Workshop) which was adapted from the basic
network and fabcar example in the [Fabric Samples](https://github.com/hyperledger/fabric-samples).
This example does not use TLS, which means the Fabric Go SDK cannot be
used to interact with the network. Currently, there are two orgs, one peer,
one orderer, and one fabric-ca in the network.

## Configuration
- [core-fpc.yaml](network-config/core-fpc.yaml) : Peer configuration that has
the SGX plugins and locations relative to location within docker image
- [core.yaml](network-config/core.yaml) : Regular Peer configuration
without FPC. Used if `$USE_FPC` is set `false`.
- [orderer.yaml](network-config/orderer.yaml) : Orderer configuration
- [crypto-config.yaml](network-config/crypto-config.yaml) : File used with cryptogen to generate
certs for specified number of orgs, peers, users, and orderer. The CA credentials
can be used to start instances of fabric-ca
- [configtx.yaml](network-config/configtx.yaml)  : File used with configtxgen to generate the
genesis block which is used as the basis of the specified channel
- [docker-compose.yml](network-config/docker-compose.yml) : Configuration of the
fabric network to be used with `docker-compose`. This file depends on two
environment variables to properly bring up a network. `$FPC_CFG` can be set to
`-fpc` or shall be empty. If set to `-fpc` the `core-fpc.yaml` & FPC peer image
is used. Otherwise it will use `core.yaml` and the regular peer image.
`$PEER_CMD` must also be set to the location of binary or script that will start
 the peer.  **Docker version 17.06.2-ce or higher is needed**

## Steps
1. Build the peer image in `utils/docker/peer` directory which is defined by the
peer [Dockerfile](../docker/peer/Dockerfile). This step
assumes you have already built the [fabric-private-chaincode base image](../docker/base/Dockerfile).
Take a look at building the docker dev environment in the main [README](../../README.md#docker).
After you have created the base image, run the following to create a modified
peer image and the plugins necessary to start the peer.  `$FPC_PATH` is the
location fabric-private-chaincode repository on your host machine.
```
cd $FPC_PATH/utils/docker/peer
docker build -t hyperledger/fabric-peer-fpc .
```
By default the image will clone the master branch on
https://github.com/hyperledger-labs/fabric-private-chaincode. If you want to use
a different fork of the repo or a different branch you provide
`FPC_REPO_URL` and `FPC_REPO_BRANCH` as build args.
```
cd $FPC_PATH/utils/docker/peer
docker build -t hyperledger/fabric-peer-fpc --build-arg FPC_REPO_URL=<repo-url> --build-arg FPC_REPO_BRANCH=<repo-branch> .
```
If you want to build the peer image using your local copy of your repo you can
use the same build args, but specify `file:///tmp/build-src/.git` as the
`FPC_REPO_URL`. You will also need to create the image at the root of this repo
so that the local repo will be in the build context for the docker daemon.
```
cd $FPC_PATH
docker build -t hyperledger/fabric-peer-fpc -f utils/docker/peer/Dockerfile --build-arg FPC_REPO_URL=file:///tmp/build-src/.git --build-arg FPC_REPO_BRANCH=$(git rev-parse --abbrev-ref HEAD) .
```
2. Download the necessary fabric binaries. Run the
[bootstrap script](scripts/bootstrap.sh) which will download the Fabric 1.4.3
into a local bin directory. The bootstrap script will download all the binaries
to the location from where the scripts are run from. The rest of the tutorial
expects the binaries to be in. If you already have the binaries downloaded in
your path, this step can be skipped.**Fabric 1.4.3 versions of configtxgen and
cryptogen are required to use the configurations above.**
```
cd $FPC_PATH/utils/docker-compose
scripts/bootstrap.sh
```

3. Generate the cryptographic material needed for the network by running the
[generate](scripts/generate.sh) script. Cryptogen will be used to generate all the
credentials needed based on the configuration filesabove and place them in the
`network-config/crypto-config` directory.  Configtxgen will be used to create
the genesis block which is used to start up the orderer as well as the peer
create channel configuration transaction. These will be placed in the
`network-config/config` directory. The `crypto-config` & `config` directory will
be mounted into every container of the FPC network as specified in the
docker-compose file. **This script is not
idempotent and will delete the contents of `crypto-config` & `config` when run
to ensure a clean start.**
```
scripts/generate.sh
```

4. Start the network. Run the [start](scripts/start.sh) script. This will use
docker-compose to start the network as well as starting the channel `mychannel`.
By default, this script will use FPC peers. If non FPC peers are desired, set
`$USE_FPC` to `false`.
```
scripts/start.sh
```

## Deploying your FPC Chaincode
The [examples](../../examples) directory has been [mounted](base/base.yaml) into
 the peer container for convenience, under
 `/project/src/github.com/hyperledger-labs/fabric-private-chaincode/examples`.
 **NOTE** If you are running a normal fabric network, the rest of the tutorial
 will not work.

1. Follow the [steps](../../examples/README.md) in the tutorial to build your
chaincode outside of the peer container. Do not continue to the testing step.
Though this tutorial references the hello world example, users can also deploy
other FPC examples using similar steps.

The rest of these steps should be done within the peer container.

2. Exec into the peer container.
```
docker exec -it peer0.org1.example.com bash
```

3. Set environment variable to use the admin credentials and set the
orderer address.
```
export CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp
export ORDERER_ADDR=orderer.example.com:7050
```

4. Install your chaincode. `PEER_CMD` is defined in the container already.
```
${PEER_CMD} chaincode install -l fpc-c -n helloworld_test -v 0 -p examples/helloworld/_build/lib
```

5. Instantiate your chaincode
```
${PEER_CMD} chaincode instantiate -o orderer.example.com:7050 -C mychannel -n helloworld_test -v 0 -c '{"Args":["init"]}' -V ecc-vscc
```

## Interact with the FPC Chaincode
1. Store asset1 with a value of a 100
```
${PEER_CMD} chaincode invoke -C mychannel -n helloworld_test -c '{"Args":["storeAsset","asset1","100"]}'
```

2. Retrieve the current value of asset1.
```
${PEER_CMD} chaincode query -C mychannel -n helloworld_test -c '{"Args":["retrieveAsset","asset1"]}'
```
The response should look like the following:
```
{
    "ResponseData":"YXNzZXQxOjEwMA==",
    "Signature":<signature>,
    "PublicKey": <public-key>
}
```

3. Verify the encrypted response data shows that asset1 is equal to a hundred.
```
> echo "YXNzZXQxOjEwMA==" | base64 -d
asset1:100
```

## Create a User with Fabric-CA
5. Enter into the [`node-sdk`](node-sdk) directory, to use the node sdk scripts
to create new users.
```
cd node-sdk
```

4. Ensure you have all the node modules
```
npm install
```

6. Enroll as the admin download the admin credentials
```
node enrollAdmin.js
```
After running this, the directory `wallet/admin` should have been created and
have public and private key pair. **NOTE** These credentials are not an admin in
the network, but just the admin for Fabric-CA and have the ability to register
more users.

7. Register another user and download the credentials.
```
node registerUser.js <username>
```
After running this with your desired username, the directory `wallet/<username>`
should have been created and have the public and private key pair.

## Interact with the chaincode using the Node SDK
**NOTE: You must run peer invoke for this chaincode once using the peer cli
commands in the peer container before you can use these node sdk scripts**

1. Ensure you have all the node modules
```
npm install
```

2. Query the asset you stored previously
```
node query.js <username> mychannel helloworld_test retrieveAsset asset1
```
The response should look similar to what you saw above when you queried using
the peer cli.
```
Transaction has been submitted, result is:
{
      "ResponseData":"YXNzZXQxOjEwMA==",
      "Signature":<signature>,
      "PublicKey":<public-key>
}
```
In general the query script works as:
```
node query.js <identity-to-use> <channel-name> <chaincode-id> <args>...
```
3. To invoke a transaction:
```
node invoke.js <username> mychannel helloworld_test storeAsset asset2 200
```
The response should look like the following:
```
Transaction has been submitted, result is:
{
      "ResponseData":"T0s=",
      "Signature":<signature>,
      "PublicKey":<public-key>
}
```
In general the invoke script works as:
```
node invoke.js <identity-to-use> <channel-name> <chaincode-id> <args>...
```

## Teardown the network

1. Run the [teardown script](./scripts/teardown.sh) to clean up your environment. Run
this in the root of this repo. **NOTE** This will try to remove all your
containers and prune all excess volumes.
```
scripts/teardown.sh
