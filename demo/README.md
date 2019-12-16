## Goal: Bring up Clock Auction Demo Application

###  Prerequisites
- It is assumed that Fabric Private Chaincode repository on your machine is installed in $FPC_PATH.
- JSON processor, jq is installed.

Instructions to bring up end-to-end application to be added here.

## Backend client

### Usage
```
# Build image once
cd $FPC_PATH/demo/client/backend
docker build -t auction_client_backend .
```

###  Register users

Before the following steps can be run, an FPC network should be setup using [README](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/utils/docker-compose/README.md).  Make sure to modify [config.json](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/demo/client/backend/fabric-gateway/config.json) to refer to the correct chaincode name.  


Register auction application users (bidders and auctioneers) with Certificate Authority
```
cd $FPC_PATH/demo/client/backend
./registerUsers.sh
```

### Run client

Note: This is work in progress.  Some environment variables are hardcoded in the following instructions since these are standalone instructions to bring up the backend client only.  Their dependencies are noted below.  As the demo application evolves, these dependencies may change.

- NETWORK_NAME:  Depends on the network setup; Refer to the usage of environment variable $USE_FPC in [README](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/utils/docker-compose/README.md).  The requirement is that the backend client runs in the _same_ network as the Fabric network.

- BACKEND_PORT:  Port at which the backend client server is available;  Default value:  3000;  Set in [config.json](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/demo/client/backend/fabric-gateway/config.json)


```                 
cd $FPC_PATH/demo/client/backend
export NETWORK_NAME=fabricfpc-fpc_basic
export BACKEND_PORT=3000
docker run --network $NETWORK_NAME -d  -v ${PWD}:/usr/src/app \
		 -p $BACKEND_PORT:$BACKEND_PORT --name client_backend auction_client_backend
```

### Test

With FPC network running and chaincode installed, you can submit transactions using curl commands.  In the same folder, `$FPC_PATH/demo/client/backend`, run:
```
./testClient.sh
```

If you see json responses for the curl commands, then connectivity to client and chaincode is verified.


### Backend apis

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
