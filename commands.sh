# ############################## Commands in order #######################################


# #################Inside the $FPC_PATH

# #  make -C $FPC_PATH/utils/docker pull pull-dev
# ### which translates to:
#     docker  pull ghcr.io/hyperledger/fabric-private-chaincode-ccenv:main
#     docker  image tag ghcr.io/hyperledger/fabric-private-chaincode-ccenv:main hyperledger/fabric-private-chaincode-ccenv:main
#     docker  pull ghcr.io/hyperledger/fabric-private-chaincode-base-dev:main
#     docker  image tag ghcr.io/hyperledger/fabric-private-chaincode-base-dev:main hyperledger/fabric-private-chaincode-base-dev:main

# #  make -C $FPC_PATH/utils/docker build build-dev
# ### which translates to:
#     docker  build --build-arg FPC_VERSION=main  -t hyperledger/fabric-private-chaincode-base-rt:main base-rt
#     docker  build --build-arg FPC_VERSION=main -t hyperledger/fabric-private-chaincode-ccenv:main ccenv
#     docker  build --build-arg FPC_VERSION=main  -t hyperledger/fabric-private-chaincode-base-dev:main base-dev

# #  make -C $FPC_PATH/utils/docker run-dev
# ### which translates to:
#     docker  build --build-arg FPC_VERSION=main -t hyperledger/fabric-private-chaincode-dev:main -f ./utils/docker/dev/Dockerfile . 


#     docker  run --rm -v "/var/run/docker.sock":"/var/run/docker.sock" -v "/src/github.com/hyperledger/fabric-private-chaincode":/project/src/github.com/hyperledger/fabric-private-chaincode --env DOCKERD_FPC_PATH=/src/github.com/hyperledger/fabric-private-chaincode/ --net=host --env SGX_MODE=SIM -i -e CI=true --name fpc-development-main -t hyperledger/fabric-private-chaincode-dev:main

# #################Inside the fpc container
#     docker  build --build-arg FPC_VERSION=main  -t hyperledger/fabric-private-chaincode-base-rt:main base-rt

#     docker  build --build-arg FPC_VERSION=main -t hyperledger/fabric-private-chaincode-ccenv:main ccenv


#     # There was make build command which was very long to right it down here 
#     # cd $FPC_PATH
#     cd samples/chaincode/cc-tools-demo && go get && cd $FPC_PATH
#     go mod vendor 
#     # make docker
#     # make build

# ################################### For CC-tools inside the fpc

# - Copy all chaincode files using the same way as simple-asset-go
# - Run `go get` inside the cc-tools-demo folder after putting it inside the fpc repo
# ------ There are huge problem with using FPC outside the FPC repository. Even go get doesn't work and you need to specify a certain version and there are conflicting packages-----------------'
# - You have to add dummy implementation for the PurgePrivate data method in the MockStup of cc-tools but be careful you need to do it in the package installed inside the FPC dev env not your local
# 	For example do: vim /project/pkg/mod/github.com/hyperledger-labs/cc-tools@v1.0.0/mock/mockstub.go and add this:
# 	// PurgePrivateData ...
# 	func (stub *MockStub) PurgePrivateData(collection, key string) error {
# 		return errors.New("Not Implemented")
# 	}
# - replace the "CHAINCODE_ID" env with "CHAINCODE_PKG_ID" in main.go

cd $FPC_PATH/samples/deployment/test-network
docker compose down

cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
./network.sh down
docker system prune
sleep 5

cd $FPC_PATH/samples/chaincode/cc-tools-demo/
# #   nano $FPC_PATH/vendor/github.com/hyperledger-labs/cc-tools/mock/mockstub.go
# # 	// PurgePrivateData ...
	# func (stub *MockStub) PurgePrivateData(collection, key string) error {
	# 	return errors.New("Not Implemented")
	# }
export CC_NAME=fpc-cc-tools-demo
make

# - run docker images | grep cc-tools-demo to make sure of the image 
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
make ercc-ecc-start
sleep 5

# prepare connections profile
cd $FPC_PATH/samples/deployment/test-network
./update-connection.sh

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
./fpcclient invoke createNewLibrary "{\"name\":\"samuel\"}"
sleep 5
./fpcclient invoke createAsset "{\"asset\":[{\"@assetType\":\"person\",\"id\":\"51027337023\",\"name\":\"samuel\"}]}"
sleep 5
./fpcclient invoke createAsset "{\"asset\":[{\"@assetType\":\"book\", \"title\": \"Fairy tail\"  ,\"author\":\"Martin\",\"currentTenant\":{\"@assetType\": \"person\", \"@key\": \"person:f6c10e69-32ae-5dfb-b17e-9eda4a039cee\"}}]}"
sleep 5
# ./fpcclient invoke getBooksByAuthor "{\"name\":\"samuel\"}" --> Fails as GetQueryResult is not implemented. I tried to implement it but the fabric implementation needs what's called handler and it's not there


# ################################# Now how to run and test cc-tools????????????????????????????????


# # -     cd ./ccapi; docker-compose up -d; cd ..





