#!/bin/bash
SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
. ${SCRIPTDIR}/common.sh

set -xe

# make sure that the docker images dev-jdoe-tlcc-0 and dev-jdoe-ecc-0 are vailable

# cd ~/fabric-v1.2

# clean up
    # rm -rf /tmp/hyperledger/*

# start orderer
    # ORDERER_GENERAL_GENESISPROFILE=SampleDevModeSolo ./build/bin/orderer

# start peer
    # ./build/bin/peer node start

# core.yaml does not understand environment variables, hence paths are relative to fabric/sgxconfig,
# so make sure we always start peer from that location, regardless where script is invoked
cd ${GOPATH}/src/github.com/hyperledger-labs/fabric-secure-chaincode/fabric/sgxconfig

peer=${FABRIC_BIN_DIR}/peer
chanid=mychannel
ccid=ecc
orderer=localhost:7050

num_rounds=3
num_clients=10

# SETUP
#============

# create channel
# - create genesis block, only by one peer
$peer channel create -o $orderer -c $chanid -f mychannel.tx
# - every peer will have to join (after having received mychannel.block out-of-band)
$peer channel join -b mychannel.block
# - every peer's tlcc will have to join as well
#   IMPORTANT: right now a join is _not_ persistant, so on restart of peer,
#   it will re-join old channels but tlcc will not!
$peer chaincode query -n tlcc -c '{"Args": ["JOIN_CHANNEL"]}' -C $chanid
sleep 3

# ercc
# - install, once per peer
$peer chaincode install -n ercc -v 0 -p github.com/hyperledger-labs/fabric-secure-chaincode/ercc
sleep 1
# - instantiate, once per channel, by single peer/admin
$peer chaincode instantiate -n ercc -v 0 -c '{"args":["init"]}' -C $chanid -V ercc-vscc
sleep 3
$peer chaincode query -n ercc -c '{"args":["getSPID"]}' -C $chanid
sleep 3


# Auction Chaincode
#=======================
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
