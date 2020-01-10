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
To set up all components at once run the start script.
```
scripts/startFPCAuctionNetwork.sh
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
in the FPC Network scripts. If you run with the `--clean-slate` flag the script
will delete all the unused volumes and chaincode images.

## Manually Bring Up The Components

### 1. FPC Network
This demo requires the use of a FPC Network. Follow the [instructions](../utils/docker-compose/README.md#Steps)
in this repo to bring one up. The demo directory is [mounted](../utils/docker-compose/network-config/docker-compose.yml)
into the peer to make installing chaincode easy.

### 2. Install the Auction Chaincode
Install the mock golang chaincode for the demo.

To install the chaincode you can run [installCC script](scripts/installCC.sh)
```
cd $FPC_PATH/demo
scripts/installCC.sh
```
If you prefer to install it manually use the following steps.

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
