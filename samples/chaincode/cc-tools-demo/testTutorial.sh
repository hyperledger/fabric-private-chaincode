#!/bin/bash

# ############################## Commands in order #######################################


# ################################### For CC-tools inside the fpc

# - Copy all chaincode files using the same way as simple-asset-go
# - replace the "CHAINCODE_ID" env with "CHAINCODE_PKG_ID" in main.go
# - Run `go get` inside the cc-tools-demo folder after putting it inside the fpc repo
# ------ There are huge problem with using FPC outside the FPC repository. Even go get doesn't work and you need to specify a certain version and there are conflicting packages-----------------'
# # - You have to add dummy implementation for the PurgePrivate data method in the MockStup of cc-tools but be careful you need to do it in the package installed inside the FPC dev env not your local
# # 	For example do: vim /project/pkg/mod/github.com/hyperledger-labs/cc-tools@v1.0.0/mock/mockstub.go and add this:
# # 	// PurgePrivateData ...
# # 	func (stub *MockStub) PurgePrivateData(collection, key string) error {
# # 		return errors.New("Not Implemented")
# # 	}.
# # A good idea is to use go mod vendor and download all go packages in the vendor directory and edit it one time there. 
# # nano $FPC_PATH/vendor/github.com/hyperledger-labs/cc-tools/mock/mockstub.go
# # 	// PurgePrivateData ...
# # 	func (stub *MockStub) PurgePrivateData(collection, key string) error {
# # 		return errors.New("Not Implemented")
# # 	}

cd $FPC_PATH/samples/deployment/test-network
docker compose down

cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
./network.sh down
docker system prune
sleep 5

cd $FPC_PATH/samples/chaincode/cc-tools-demo/
export CC_NAME=fpc-cc-tools-demo
make

# - run docker images | grep fpc-cc-tools-demo to make sure of the image 
# - complete the tutorial normally:
cd $FPC_PATH/samples/deployment/test-network
./setup.sh

cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
./network.sh up createChannel -ca -c mychannel
sleep 5

export CC_ID=cc-tools-demo
export CC_PATH="$FPC_PATH/samples/chaincode/cc-tools-demo/"
export CC_VER=$(cat "$FPC_PATH/samples/chaincode/cc-tools-demo/mrenclave")

cd $FPC_PATH/samples/deployment/test-network
./installFPC.sh
sleep 5
export EXTRA_COMPOSE_FILE="$FPC_PATH/samples/chaincode/cc-tools-demo/cc-tools-demo-compose.yaml"
make ercc-ecc-start
sleep 5

# # prepare connections profile
cd $FPC_PATH/samples/deployment/test-network
./update-connection.sh

# update the connection profile for external clients outside the fpc dev environment
cd $FPC_PATH/samples/deployment/test-network
./update-external-connection.sh

# make fpcclient
cd $FPC_PATH/samples/application/simple-cli-go
make

# export fpcclient settings
export CC_NAME=cc-tools-demo
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

sleep 5
# init our enclave
./fpcclient init $CORE_PEER_ID
sleep 5
# invoke the getSchema transaction which is implemented internally by cc-tools
./fpcclient invoke getSchema

########################## Some transactions to test ####################################

##NOTE: In cc-tools-demo, most of these transactions set permissions to filter which orgs are allowed to invoke it or not. The current organization used in this script is "Org1MSP".
## Beware that org names are case sensitive 

# sleep 5
# ./fpcclient invoke createNewLibrary "{\"name\":\"samuel\"}"
# sleep 5
# ./fpcclient invoke createAsset "{\"asset\":[{\"@assetType\":\"person\",\"id\":\"51027337023\",\"name\":\"samuel\"}]}"
# sleep 5
# ./fpcclient invoke createAsset "{\"asset\":[{\"@assetType\":\"book\", \"title\": \"Fairy tail\"  ,\"author\":\"Martin\",\"currentTenant\":{\"@assetType\": \"person\", \"@key\": \"person:f6c10e69-32ae-5dfb-b17e-9eda4a039cee\"}}]}"
# sleep 5
# ./fpcclient invoke getBooksByAuthor "{\"authorName\":\"samuel\"}" # --> Fails as GetQueryResult is not implemented. I tried to implement it but the fabric implementation needs what's called handler and it's not ther


