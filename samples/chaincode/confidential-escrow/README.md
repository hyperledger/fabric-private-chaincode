# Runnign Procedure

> 1. In 1st terminal Window

```bash
make -C $FPC_PATH/utils/docker run-dev

GOOS=linux make -C $FPC_PATH/ercc build docker
GOOS=linux make -C $FPC_PATH/samples/chaincode/confidential-escrow with_go docker

cd $FPC_PATH/samples/chaincode/confidential-escrow
# make

# cd $FPC_PATH/samples/deployment/test-network # 1 time
# ./setup.sh # 1 time

cd $FPC_PATH/samples/deployment/test-network/fabric-samples/test-network
./network.sh down
./network.sh up -ca
./network.sh createChannel -c mychannel

export CC_ID=confidential-escrow
export CC_PATH="$FPC_PATH/samples/chaincode/confidential-escrow/"
export CC_VER=$(cat "$FPC_PATH/samples/chaincode/confidential-escrow/mrenclave")
cd $FPC_PATH/samples/deployment/test-network
./installFPC.sh

export EXTRA_COMPOSE_FILE="$FPC_PATH/samples/chaincode/confidential-escrow/confidential-escrow-compose.yaml"
make ercc-ecc-start
```

> 2. In 2nd terminal window

```bash
docker exec -it fpc-development-main /bin/bash

# prepare connections profile
cd $FPC_PATH/samples/deployment/test-network
./update-connection.sh

# # update the connection profile for external clients outside the FPC dev environment
cd $FPC_PATH/samples/deployment/test-network
./update-external-connection.sh

# make fpcclient
cd $FPC_PATH/samples/application/simple-cli-go
make

# export fpcclient settings
export CC_ID=confidential-escrow
export CHANNEL_NAME=mychannel
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_ID=peer0.org1.example.com
export CORE_PEER_ORG_NAME=org1
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_MSPCONFIGPATH=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_TLS_CERT_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt
export CORE_PEER_TLS_ENABLED="true"
export CORE_PEER_TLS_KEY_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
export CORE_PEER_TLS_ROOTCERT_FILE=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export ORDERER_CA=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export GATEWAY_CONFIG=$FPC_PATH/samples/deployment/test-network/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml
export FPC_ENABLED=true
export RUN_CCAAS=true

# init our enclave
./fpcclient init $CORE_PEER_ID
```

> 3. Run transaction

```bash
./fpcclient invoke getSchema

./fpcclient invoke debugTest '{}'

./fpcclient invoke createDigitalAsset '{
  "name": "CBDC",
  "symbol": "CBDC",
  "decimals": 2,
  "totalSupply": 1000000,
  "owner": "central_bank",
  "issuerHash": "sha256:abc123"
}'

./fpcclient invoke createWallet '{
  "walletId": "wallet-123",
  "ownerId": "Abhinav",
  "ownerCertHash": "sha256:def456",
  "balance": 0,
  "digitalAssetType": "CBDC"
}'

./fpcclient invoke createEscrow '{
  "escrowId": "escrow-456",
  "buyerPubKey": "buyer_pub",
  "sellerPubKey": "seller_pub",
  "amount": 1000,
  "assetType": "CBDC",
  "conditionValue": "sha256:secret123",
  "status": "Active",
  "buyerCertHash": "sha256:buyer_cert"
}'
```
