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

## Starting the network

### Quick start
   The quickest way to get up and running is to simply execute
   ```
   scripts/start.sh
   ```
   This will create all necessary installation artifacts and start the
   network. 
   If your environment variable `SGX_MODE` is set to hardware, the network will run
   the peer also with SGX hardware mode enabled, otherwise it will run in SGX simulation mode.
   If you set the environment variable `USE_EXPLORER` to `true`, the network will include
   and start the [Hyperledger Explorer](https://www.hyperledger.org/projects/explorer) on 
   [port 8090](http://localhost:8090). This will enable you to inspect the networks, 
    e.g., processed transactions.
   If you set the environment variable `USE_COUCHDB` to `true`, the peer will use couchdb
   to store the local version of the ledger and you can inspect the peer's ledger state
   on [port 5984](http://localhost:5984) (login as user `admin` with password `adminPassword`).

   For more information in the steps involved, continue
   reading the following section. Otherwise, you can skip to the
   Section on [Chaincode Installation](#deploying-your-fpc-chaincode).


### Detailed Steps
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
   `FPC_REPO_URL` and `FPC_REPO_BRANCH_TAG_OR_COMMIT` as build args.
   ```
   cd $FPC_PATH/utils/docker/peer
   docker build -t hyperledger/fabric-peer-fpc --build-arg FPC_REPO_URL=<repo-url> --build-arg FPC_REPO_BRANCH_TAG_OR_COMMIT=<repo-branch> .
   ```
   If you want to build the peer image using your local copy of your repo you can
   use the same build args, but specify `file:///tmp/build-src/.git` as the
   `FPC_REPO_URL`. You will also need to create the image at the root of this repo
   so that the local repo will be in the build context for the docker daemon.
   ```
   cd $FPC_PATH
   docker build -t hyperledger/fabric-peer-fpc -f utils/docker/peer/Dockerfile --build-arg FPC_REPO_URL=file:///tmp/build-src/.git --build-arg FPC_REPO_BRANCH_TAG_OR_COMMIT=$(git rev-parse HEAD) .
   ```
   Note: as this last scenario might be a common development action, it is defined as
   a makefile target `peer` in `$FPC_PATH/utils/docker/Makefile`.

2. Download the necessary fabric binaries. Run the
   [bootstrap script](scripts/bootstrap.sh) which will download the Fabric 1.4.3
   binaries into `$FPC_PATH/utils/docker-compose/bin` directory as well as download
   also all fabric docker images  that version. If you already have the binaries
   downloaded and in located in your `PATH`, this step can be skipped. If you don't want
   download the docker images, you can also skip that part by passing option `-d` to
   the `bootstrap.sh` script (docker-compose will download them later if they are missing
   locally.)
   **Fabric 1.4.3 versions of configtxgen and  are required to use the configurations above.**
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
   **Note**
   - if some of steps 1 to 3 were omitted before running start.sh, the
     script will perform the missing steps in the default configuration
   - the script returns to you an export statements with environment variables
     which enable you to easily run `docker-compose` commands such as `ps`, `top`, `logs`
     and alike. Just copy/paste the export statement into your shell and you can get,
     e.g., the container status with `${DOCKER_COMPOSE} ps`.

## Deploying your FPC Chaincode
The [examples](../../examples) and [demo](../../demo) directories has been
[mounted](base/base.yaml) into the peer container for convenience, under
`/project/src/github.com/hyperledger-labs/fabric-private-chaincode/examples` and
`/project/src/github.com/hyperledger-labs/fabric-private-chaincode/demo`.
 **NOTE** If you are running a normal fabric network, the rest of the tutorial
 will not work.

1. Follow the [steps](../../examples/README.md) in the tutorial to build your
   chaincode outside of the peer container. Do not continue to the testing step.
   Though this tutorial references the hello world example, users can also deploy
   other FPC examples, e.g., the [echo](../../examples/echo) or
   [auction](../../examples/auction) examples, where the code is provided in the git repo out-of-the-box.
   Follow similar steps as outlined below with corresponding changes of chaincode name and
   queries/transactions.
   
   The rest of these steps should be done within the peer container.

2. Exec into the peer container.
   ```
   docker exec -it peer0.org1.example.com bash
   ```

3. There also some a number of useful predefined environment variables such as the orderer address, 
   the channel name, the peer command to use and the credentials used by fabric.
   ```
   echo CORE_PEER_MSPCONFIGPATH=${CORE_PEER_MSPCONFIGPATH}\
        ORDERER_ADDR=${ORDERER_ADDR}\
        CHANNEL_NAME=${CHANNEL_NAME}\
        PEER_CMD=${PEER_CMD}\
        CORE_PEER_MSPCONFIGPATH=${CORE_PEER_MSPCONFIGPATH}
   ```
   Note that to save you from explicitly having to specify CORE_PEER_MSPCONFIGPATH on each call, we defined them
   in `docker-compose.yml` as admin credentials. This means though also that your peer will run with admin
   instead of peer credentials. If you have role-specific endorsement policies, you might have to comment out the
   corresponding definition in [docker-compose.yml](./network-config/docker-compose.yml) to peer credentials and then manually
   define the corresponding value here in the docker exec shell.

4. Install your chaincode.
   ```
   ${PEER_CMD} chaincode install -l fpc-c -n helloworld_test -v 0 -p examples/helloworld/_build/lib
   ```

5. Instantiate your chaincode
   ```
   ${PEER_CMD} chaincode instantiate -C mychannel -n helloworld_test -v 0 -c '{"Args":["init"]}' -V ecc-vscc
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
1. Enter into the [`node-sdk`](node-sdk) directory, to use the node sdk scripts
   to create new users.
   ```
   cd node-sdk
   ```

2. Ensure you have all the node modules
   ```
   npm install
   ```

3. Enroll as the admin download the admin credentials
   ```
   node enrollAdmin.js
   ```
   After running this, the directory `wallet/admin` should have been created and
   have public and private key pair. **NOTE** These credentials are not an admin in
   the network, but just the admin for Fabric-CA and have the ability to register
   more users.

4. Register another user and download the credentials.
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

1. Run the [teardown script](./scripts/teardown.sh) to clean up your environment.
   To do a clean state teardown, add option `--clean-slate`.
   **NOTE** `--clean-slate` will try to remove all anonymouse volumes (which
   includes the state of the CA, the wallets in `${FPC_PATH}/utils/docker-compse/node-sdk`
   as well as _all_ your containers and left-over chaincode containers.
   ```
   scripts/teardown.sh
   ```
