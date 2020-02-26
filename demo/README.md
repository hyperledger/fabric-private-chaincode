# Clock Auction Demo Application

##  Prerequisites
- It is assumed that Fabric Private Chaincode repository on your machine is installed in $FPC_PATH.
- JSON processor, jq is installed.
- FPC Peer Docker Image has been created as per the FPC Network
[Instructions](../utils/docker-compose/README.md#Steps) in Step 1.
- Relevant Fabric binaries and docker images have been downloaded as per FPC
Network [Instructions](../utils/docker-compose/README.md#Steps) in Step 2.

The Auction Demo has multiple components, a UI, a backend server that proxies calls
to an FPC network, a FPC network and a chaincode to execute the auction logic.

## Bring Up the Demo End To End
### Setup
To set up all components at once run the start script. An [FPC Network](../utils/docker-compose/network-config/docker-compose.yml)
will automatically be created.
```
cd $FPC_PATH/demo
scripts/startFPCAuctionNetwork.sh --build-cc
```

The channel `mychannel` will be created and used to install and instantiate
all chaincodes. The [golang mock chaincode](chaincode/golang) will be
instantiated as `mockcc` and the [FPC auction chaincode](chaincode/fpc) will be
instantiated as `ecc_auctioncc`.  If you do not need to build
the FPC Auction CC, omit the `--build-cc` flag. If you do pass the `--build-cc`
flag, the script assumes that the docker image
`hyperledger/fabric-private-chaincode-cc-builder` exists. If the image does not
exist, the [cc-builder](../utils/docker/cc-builder/Dockerfile) and the
[dev](../utils/docker/dev/Dockerfile) images will be built automatically before
building the chaincode.

The fabric gateway will be configured to use the `auctioncc` chaincode. This can
be changed by changing the `chaincode_name` in the fabric gateway
[config](client/backend/fabric-gateway/config.json). Currently it is set to
`ecc_auctioncc`. If you change the fabric gateway config, remember to rebuild
the client so the updated configuration is added in the fabric gateway image.
You can add the `--build-client` flag to the above start script to automatically
rebuild the fabric gateway and frontend.

**NOTE** To differentiate FPC chaincode from other chaincodes,
we currently prefix the chaincode name with `ecc_` when installing it. While the
FPC peer cli hides this name mapping, you will have to manually prefix the
chaincode name in `config.json`, hence the default `ecc_auctioncc`.

Both the frontend and fabric-gateway [expose ports](docker-compose.yml) and are
accessible on the host machine. The frontend can be accessed at [localhost:5000](http://localhost:5000)
and the fabric-gateway can be accessed at [localhost:3000](http://localhost:3000).

Below is the script's help text.

```
startFPCAuctionNetwork.sh [options]

   This script, by default, will teardown possible previous iterations of this
   demo, generate new crypto material for the network, start an FPC network as
   defined in \$FPC_PATH/utils/docker-compose, install the mock golang auction
   chaincode(\$FPC_PATH/demo/chaincode/golang), install the FPC compliant
   auction chaincode(\$FPC_PATH/demo/chaincode/fpc), register auction users,
   and bring up both the fabric-gatway & frontend UI.

   If the fabric-gateway and frontend UI docker images have not previously been
   built it will build them, otherwise the script will reuse the images already
   existing.  You can force a rebuild, though, by specifying the flag
   --build-client.  The FPC chaincode will not be built unless specified by the
   flag --build-cc.  By calling the script with both build options, you will be
   able to run the demo without having to build the whole FPC project (e.g., by
   calling `make` in $FPC_PATH).

   options:
       --build-cc:
           As part of bringing up the demo components, the auction cc in demo/chaincode/fpc will
           be rebuilt using the docker-build make target.
       --build-client:
           As part of bringing up the demo components, the Fabric Gateway and the UI docker images
           will be built or rebuilt using current source code.
       --help,-h:
           Print this help screen.
```
**NOTE** The above [script](scripts/startFPCAuctionNetwork.sh) will bring up a
fresh FPC Network and generate new credentials using the FPC Network Setup
[scripts](../utils/docker-compose/scripts). Therefore this script should only be
run when no other fabric network is running to avoid port collisions.

Both the frontend and fabric-gateway [expose ports](docker-compose.yml) and are
accessible on the host machine. The frontend can be accessed at `localhost:5000`
and the fabric-gateway can be accessed at `localhost:3000`.

### Teardown
To bring down all of the components and the underlying FPC network run the
following script.
```
scripts/teardown.sh
```
**NOTE** The script will run the [teardown script](../utils/docker-compose/scripts/teardown.sh)
in the FPC Network scripts. If you run it with the `--clean-slate` flag the script
will delete all the unused volumes and chaincode images.


### Scripting

To facilitate demonstrations and also to help in testing, you can specify with a simple
[DSL](client/scripting/lib/dsl.sh) a scenario script defining the
actions of the different parties and execute it using the command
[scenario-run.sh](client/scripting/scenario-run.sh).  
Below is the script's help text.
```
scenario-run.sh [--help|-h|-?] [--bootstrap|-b] [--dry-run|-d] [--non-interactive|-n] [--skip-delay|-s] [--mock-reset|-r] <script-file>
    Run the demo scenario codified in the passed script file.
    - If you pass option --bootstrap, it will also first bring up an FPC network
      and tear it down at the end; otherwise it assumes you have already
      a running setup ...
    - option --dry-run/-d will just display all requests but not execute/submit
      any of them
    - option --non-interactive/-n will case _all_ requests to be submited even
      requests from submit_manual.  This allows you to easily validate all
      json files, even when some steps would be manual in an actual scenario
    - option --skip-delay/-s allows you ignore all delays to speed-up demo
    - option --mock-reset/-r will try reset the mock-server at the beginning
      (obviously, this won't work if you run against the fabric-gateway backend;
       to achieve the equivalent for fabric-gateway, use option --bootstrap)
```

An example of a scenario can be found in [demo/scenario](scenario).

## Manually Bring Up The Components

### 1. FPC Network
This demo requires the use of a FPC Network. Follow the [instructions](../utils/docker-compose/README.md#Steps)
in this repo to bring one up. The demo directory is [mounted](../utils/docker-compose/network-config/docker-compose.yml)
into the peer to make installing chaincode easy.

### 2. Install the Auction Chaincode

#### Build the FPC Auction Chaincode
Build the FPC Auction Chaincode in a docker image. The make target makes use
of the `hyperledger/fabric-private-chaincode-cc-builder`. If the image does not
already exist, the target will build the [cc-builder](../utils/docker/cc-builder/Dockerfile)
and the [dev](../utils/docker/dev/Dockerfile) images automatically before
building the chaincode.
```
cd $FPC_PATH/demo/chaincode/fpc
make docker-build
```
If you do not wish to use docker to build the chaincode you can build directly.
The FPC project must be built to be able to run this.
```
cd $FPC_PATH/demo/chaincode/fpc
make build
```

#### Install the FPC Auction Chaincode
Install the mock golang chaincode for the demo.

To install the chaincode you can run [installCC script](scripts/installCC.sh)
```
cd $FPC_PATH/demo
scripts/installCC.sh
```
If you prefer to install it manually use the following steps.

1. Exec into the peer container.The demo directory has been mounted into the
peer container for convenience, so the chaincode build files will be available.
There are environment variables already set in the peer container that will be
convenient for next set of steps. Please refer to the FPC Network Setup
[documentation](../utils/docker-compose/README.md#Deploying-your-FPC-Chaincode)
for details on what environment variables exist and how to see their values.
```
docker exec -it peer0.org1.example.com
```

2. Install the Auction Chaincode
```
${PEER_CMD} chaincode install -n auctioncc -v 1.0 --path demo/chaincode/fpc/_build/lib -l fpc-c
```

3. Instantiate the Auction chaincode
```
${PEER_CMD} chaincode instantiate -n auctioncc -v 1.0 --channelID mychannel -c '{"Args":[]}'
```

4. Exit the peer container
```
ctrl + d
```

#### Install Mock Chaincode

1. Exec into the peer container. There are environment variables already set in
the peer container that will be convenient for next set of steps. Please refer
to the FPC Network Setup [documentation](../utils/docker-compose/README.md#Deploying-your-FPC-Chaincode)
for details on what environment variables exist and how to see their values.
```
docker exec -it peer0.org1.example.com bash
```

2. Install the mockcc.
```
${PEER_CMD} chaincode install -n mockcc -v 1.0 --path github.com/hyperledger-labs/fabric-private-chaincode/demo/chaincode/golang/cmd -l golang
```

3. Instantiate the mockcc
```
${PEER_CMD} chaincode instantiate -n mockcc -v 1.0 --channelID mychannel -c '{"Args":[]}'
```

4. Exit the container
```
ctrl + d
```

### 3. Fabric Gateway
####  Register users

Before the following steps can be run, an FPC network should be setup using [README](../utils/docker-compose/README.md).  Make sure to modify [config.json](client/backend/fabric-gateway/config.json) to refer to the correct chaincode name.


Register auction application users (bidders and auctioneers) with Certificate Authority
```
cd $FPC_PATH/demo/client/backend/fabric-gateway
./registerUsers.sh
```

#### Run client

Note: This is work in progress.  Some environment variables are hardcoded in the following instructions since these are standalone instructions to bring up the backend client only.  Their dependencies are noted below.  As the demo application evolves, these dependencies may change.

- NETWORK_NAME:  Depends on the network setup; Refer to the usage of environment variable $USE_FPC in [README](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/utils/docker-compose/README.md).  The requirement is that the backend client runs in the _same_ network as the Fabric network.

- BACKEND_PORT:  Port at which the backend client server is available;  Default value:  3000;  Set in [config.json](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/demo/client/backend/fabric-gateway/config.json)


```
cd $FPC_PATH/demo
COMPOSE_PROJECT_NAME=fabricfpc docker-compose -f docker-compose.yml up -d auction_client
```

#### Test

With FPC network running and chaincode installed, you can submit transactions using curl commands.  In the same folder, `$FPC_PATH/demo/client/backend/fabric-gateway`, run:
```
./testClient.sh
```

If you see json responses for the curl commands, then connectivity to client and chaincode is verified.


#### Backend apis

Following samples illustrate the urls of the apis with the default value of $BACKEND_PORT which is 3000.

Get the list of registered users.  No authentication is in place for this api.  
```
http://localhost:3000/api/getRegisteredUsers
```

Get the default auction in json format
```
http://localhost:3000/api/clock_auction/getDefaultAuction
```

Use following functions to invoke a transaction or query.  Please note that any chaincode function can be called using invoke or query.   The api expects the following header:  `{x-user:username}` where `username` is an entry in the list from `getRegisteredUsers`.
```
http://localhost:3000/api/cc/invoke
http://localhost:3000/api/cc/query
```

### 4. Frontend UI

1. Run the docker image
```
cd $FPC_PATH/demo
COMPOSE_PROJECT_NAME=fabricfpc docker-compose -f docker-compose.yml up -d auction_frontend
```
The frontend expects that it can talk to the fabric-gateway on port 3000, so it
needs to be run on the same network as the fabric-gateway and FPC Network.

2. Navigate to `localhost:5000` in the browser.

### 5. Teardown
The [teardown script](scripts/teardown.sh) can still be used to bring down all the demo components.
```
cd $FPC_PATH/demo
scripts/teardown.sh
```
**Note** Add `--clean-slate` when running the teardown script to clear all
unused volumes and chaincode images.

If you prefer to manually bring down the components use the following steps.

1. Bring down the frontend UI & fabric-gateway
```
cd $FPC_PATH/demo
COMPOSE_PROJECT_NAME=fabricfpc docker -f docker-compose.yml down
```
2. Bring down the FPC network
```
cd $FPC_PATH/utils/docker-compose
scripts/teardown.sh
```
**Note** Add `--clean-slate` when running the teardown script to clear all
unused volumes.
