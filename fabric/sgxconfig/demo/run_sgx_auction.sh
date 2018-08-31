#!/bin/bash
set -xe

# make sure that the docker images dev-jdoe-tlcc-0 and dev-jdoe-ecc-0 are vailable

# cd ~/fabric-v1.2

# clean up
    # rm -rf /tmp/hyperledger/*

# start orderer
    # ORDERER_GENERAL_GENESISPROFILE=SampleDevModeSolo ./build/bin/orderer

# start peer
    # ./build/bin/peer node start

peer=../.build/bin/peer
chanid=mychannel
ccid=ecc
orderer=localhost:7050

num_rounds=3
num_clients=10

# create channel
$peer channel create -o $orderer -c $chanid -f mychannel.tx
$peer channel join -b mychannel.block
$peer chaincode query -n tlcc -c '{"Args": ["JOIN_CHANNEL"]}' -C $chanid
sleep 3

# ercc
$peer chaincode install -n ercc -v 0 -p github.com/hyperledger-labs/fabric-secure-chaincode/ercc
sleep 1
$peer chaincode instantiate -n ercc -v 0 -c '{"args":["init"]}' -C $chanid -V ercc-vscc
sleep 3
$peer chaincode query -n ercc -c '{"args":["getSPID"]}' -C $chanid
sleep 3

# install, init, and register (auction) chaincode
# install some dummy chaincode (we manually need to create the image)
$peer chaincode install -n $ccid -v 0 -p github.com/hyperledger/fabric/examples/chaincode/go/example02/cmd 
sleep 3
$peer chaincode instantiate -o $orderer -C $chanid -n $ccid -v 0 -c '{"args":["init"]}' -V ecc-vscc
sleep 3

$peer chaincode invoke -o $orderer -C $chanid -n $ccid -c '{"Args":["setup", "ercc"]}'
sleep 3

# $peer chaincode query -o $orderer -C $chanid -n $ccid -c '{"Args":["getEnclavePk"]}'

# create auction
$peer chaincode invoke -o $orderer -C $chanid -n $ccid -c '{"Args": ["[\"create\",\"MyAuction\"]", ""]}'
sleep 3

echo "invoke submit"
$peer chaincode invoke -o $orderer -C $chanid -n $ccid -c '{"Args":["[\"submit\",\"MyAuction\", \"JohnnyCash0\", \"0\"]", ""]}'
sleep 3

for (( i=1; i<=$num_rounds; i++ ))
do
    b="$(($i%$num_clients))"
    $peer chaincode invoke -o $orderer -C $chanid -n $ccid -c '{"Args":["[\"submit\",\"MyAuction\", \"JohnnyCash'$b'\", \"'$b'\"]", ""]}'
done
sleep 3

$peer chaincode invoke -o $orderer -C $chanid -n $ccid -c '{"Args":["[\"close\",\"MyAuction\"]",""]}'
sleep 3

echo "invoke eval"
for (( i=1; i<=1; i++ ))
do
    $peer chaincode invoke -o $orderer -C $chanid -n $ccid -c '{"Args":["[\"eval\",\"MyAuction\"]", ""]}'
done
