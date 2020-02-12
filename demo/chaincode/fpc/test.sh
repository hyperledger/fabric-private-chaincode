#!/bin/bash
#
# Copyright 2019 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

function cleanup {
    killall mock
}

function check_success {
    RC=`echo "$RESPONSE" | jq -r .status | jq -r .rc`; if [[ $RC == 0 ]]; then return 0; else return 1; fi
}

function check_failure {
    if check_success; then return 1; else return 0; fi
}

FAIL_ON_SUCCESS="if check_success ; then echo \"TEST FAILED: \${LINENO}\"; exit -1; fi"
FAIL_ON_FAILURE="if check_failure ; then echo \"TEST FAILED: \${LINENO}\"; exit -1; fi"

trap cleanup EXIT

export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:+"${LD_LIBRARY_PATH}:"}${FPC_PATH}/ecc_enclave/_build/lib

pushd ${FPC_PATH}/demo/client/backend/mock
test -e mock || echo "Compiling mock server..." && make build
rm -f enclave && ln -s ${FPC_PATH}/demo/chaincode/fpc/_build enclave
./mock 2>&1 /dev/null &
pushd
sleep 2

#(fails, no territories no bidders) create auction
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"createAuction","args":["{\"name\":\"bruno\", \"territories\":[], \"bidders\":[], \"initialEligibilities\":[], \"activityRequirementPercentage\": 0, \"clockPriceIncrementPercentage\": 1000}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(succeeds) create auction
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"createAuction","args":["{\"name\":\"bruno\", \"territories\":[{\"id\": 1, \"name\": \"myterritory\", \"isHighDemand\": false, \"minPrice\": 200, \"channels\": [{\"id\": 1, \"name\": \"mychannel1\", \"impairment\":20}, {\"id\": 2, \"name\": \"mychannel2\", \"impairment\":40}, {\"id\": 3, \"name\": \"mychannel3\", \"impairment\":60}]}, {\"id\": 2, \"name\": \"myterritory2\", \"isHighDemand\": false, \"minPrice\": 300, \"channels\": [{\"id\": 1, \"name\": \"mychannel3\", \"impairment\":33}, {\"id\": 2, \"name\": \"mychannel4\", \"impairment\":66}, {\"id\": 3, \"name\": \"mychannel5\", \"impairment\":99}]}], \"bidders\":[{\"id\": 1, \"displayName\": \"mickey\", \"principal\":{\"id\": 1, \"mspId\": \"Org1MSP\",\"dn\": \"CN=TestUser-1,OU=user+OU=org1\"}}, {\"id\": 2, \"displayName\": \"duffy\", \"principal\":{\"id\": 2, \"mspId\": \"Org1MSP\",\"dn\": \"CN=TestUser-2,OU=user+OU=org1\"}}, {\"id\": 3, \"displayName\": \"goku\", \"principal\":{\"id\": 3, \"mspId\": \"Org1MSP\",\"dn\": \"CN=TestUser-3,OU=user+OU=org1\"}}], \"initialEligibilities\":[{\"bidderId\": 1, \"number\": 4}, {\"bidderId\": 2, \"number\": 6}, {\"bidderId\": 3, \"number\": 6}], \"activityRequirementPercentage\": 0, \"clockPriceIncrementPercentage\": 20}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

RESPONSE=`curl -s -H "Content-Type: application/json" -X POST -d '{"tx":"getAuctionDetails", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE


RESPONSE=`curl -s -H "Content-Type: application/json" -X POST -d '{"tx":"getAuctionStatus", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(fails, empty bids) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 1, \"bids\":[]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(fails, unrecognized submitter) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 2, \"bids\":[{\"terId\": 1, \"qty\":4, \"price\":2000}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(fails, round not current) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 2, \"bids\":[{\"terId\": 1, \"qty\":4, \"price\":2000}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(fails, round not active) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 1, \"bids\":[{\"terId\": 1, \"qty\":4, \"price\":2008}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(fails, unrecognized submitter) start next round
RESPONSE=`curl -s -H "Content-Type: application/json" -X POST -d '{"tx":"startNextRound", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(succeeds) start next round
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"startNextRound", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(fails, too much demand) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-2" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 1, \"bids\":[{\"terId\": 1, \"qty\":4, \"price\":2008}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(fails, not enough eligibility) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 1, \"bids\":[{\"terId\": 1, \"qty\":3, \"price\":2000}, {\"terId\": 2, \"qty\":3, \"price\":2000}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(succeeds) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 1, \"bids\":[{\"terId\": 1, \"qty\":2, \"price\":2000}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds, overwrites previous one) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 1, \"bids\":[{\"terId\": 1, \"qty\":2, \"price\":1500}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-2" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 1, \"bids\":[{\"terId\": 1, \"qty\":3, \"price\":1588}, {\"terId\": 2, \"qty\":3, \"price\":1588}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"endRound", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(fails, round not current) submit bid 
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 1, \"bids\":[{\"terId\": 1, \"qty\":2, \"price\":1500}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(succeeds) start round
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"startNextRound", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(fails, bid above clock price) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 2, \"bids\":[{\"terId\": 1, \"qty\":2, \"price\":1500}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(fails, bid below posted price) submit bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 2, \"bids\":[{\"terId\": 1, \"qty\":2, \"price\":15}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(succeeds)submit bid to maintain demand and increase demand in territory 1,2 resp.
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 2, \"bids\":[{\"terId\": 1, \"qty\":2, \"price\":200}, {\"terId\": 2, \"qty\":2, \"price\":300}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds)submit bid to maintain demand in territory 1,2
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-2" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 2, \"bids\":[{\"terId\": 1, \"qty\":3, \"price\":208}, {\"terId\": 2, \"qty\":3, \"price\":308}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds) end round
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"endRound", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds) start round
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"startNextRound", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds)submit bid to maintain demand and increase demand in territory 1,2 resp.
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitClockBid", "args":["{\"auctionId\": 1, \"round\": 3, \"bids\":[{\"terId\": 1, \"qty\":2, \"price\":240}, {\"terId\": 2, \"qty\":2, \"price\":360}]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds) end round
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"endRound", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds) get round info
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"getRoundInfo", "args":["{\"auctionId\": 1, \"round\": 2}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(fails, round 0) get round info
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"getRoundInfo", "args":["{\"auctionId\": 1, \"round\": 0}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(succeeds, inner round format) get bidder round results
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"getBidderRoundResults", "args":["{\"auctionId\": 1, \"round\": 2}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds, last round format) get bidder round results
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"getBidderRoundResults", "args":["{\"auctionId\": 1, \"round\": 3}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(fails, not owner) get owner round results
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"getOwnerRoundResults", "args":["{\"auctionId\": 1, \"round\": 2}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(succeeds, inner round) get owner round results
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"getOwnerRoundResults", "args":["{\"auctionId\": 1, \"round\": 2}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds, last round) get owner round results
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"getOwnerRoundResults", "args":["{\"auctionId\": 1, \"round\": 3}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(fails, unimplented api) submit assign bid
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestUser-1" -X POST -d '{"tx":"submitAssignmentBid", "args":["{\"auctionId\": 1, \"bids\":[]}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(fails, assignment phase already done) start round
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:TestAuctioneer-1" -X POST -d '{"tx":"startNextRound", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_SUCCESS

#(succeeds, assignment results) get assignment results
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:anybody" -X POST -d '{"tx":"getAssignmentResults", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

#(succeeds, publish assignment results) publish assignment results
RESPONSE=`curl -s -H "Content-Type: application/json" -H "x-user:anybody" -X POST -d '{"tx":"publishAssignmentResults", "args":["{\"auctionId\": 1}"]}' http://localhost:3000/api/cc/invoke`
eval $FAIL_ON_FAILURE

echo "Test successful."

exit 0

