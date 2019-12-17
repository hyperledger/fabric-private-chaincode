#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

#!/bin/bash

#  These curl commands are provided here to illustrate the usage of the backend client.  
#  With a Fabric network running, a chaincode installed and the client deployed, this script
#  exercises all of the exposed apis of the client.  However, the inputs do not cover all 
#  possible inputs for the purpose of testing the client.  

#  Future enhancement:  This script could be extended to test the client by adding all possible
#  inputs and by adding checks to verify the return values / errors.

set -ev
curl -H "Content-Type: application/json" -X GET  http://localhost:3000/api/clock_auction/getDefaultAuction


curl -H "Content-Type: application/json" -X GET  http://localhost:3000/api/getRegisteredUsers


#  "userName": "Auctioneer1"

curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"createAuction","args":["unused"]}' http://localhost:3000/api/cc/invoke


curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"getAuctionDetails","args":["unused"]}' http://localhost:3000/api/cc/invoke


curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"getAuctionStatus","args":["unused"]}' http://localhost:3000/api/cc/invoke


curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"startNextRound","args":["unused"]}' http://localhost:3000/api/cc/invoke


curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"endRound","args":["unused"]}' http://localhost:3000/api/cc/invoke


# userName = A-Telecom;

curl -H "Content-Type: application/json" -H "x-user:A-Telecom" -X POST -d '{"tx":"submitInitialClockBid","args":["unused"]}' http://localhost:3000/api/cc/invoke


curl -H "Content-Type: application/json" -H "x-user:A-Telecom" -X POST -d '{"tx":"submitRegularClockBid","args":["unused"]}' http://localhost:3000/api/cc/invoke


#  userName = B-Net;

curl -H "Content-Type: application/json" -H "x-user:B-Net" -X POST -d '{"tx":"submitAssignBid","args":["unused"]}' http://localhost:3000/api/cc/invoke


curl -H "Content-Type: application/json" -H "x-user:B-Net" -X POST -d '{"tx":"getAssignmentResults","args":["unused"]}' http://localhost:3000/api/cc/invoke




#............  Queries  ..............

# userName = "C-Mobile"

curl -H "Content-Type: application/json" -H "x-user:C-Mobile" -X POST -d '{"tx":"getRoundInfo","args":["unused"]}' http://localhost:3000/api/cc/query


curl -H "Content-Type: application/json" -H "x-user:C-Mobile" -X POST -d '{"tx":"getBidderRoundResults","args":["unused"]}' http://localhost:3000/api/cc/query


#  "userName": "Auctioneer1"

curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"getAuctionDetails","args":["unused"]}' http://localhost:3000/api/cc/query


curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"getAuctionStatus","args":["unused"]}' http://localhost:3000/api/cc/query


curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"getOwnerRoundResults","args":["unused"]}' http://localhost:3000/api/cc/query


curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"getAssignmentResults","args":["unused"]}' http://localhost:3000/api/cc/query


# .....................................................
# ...... For following test cases;  expect error ......
# .....................................................
# Test invoke with newUser not registered with CA;  expect error

curl -H "Content-Type: application/json" -H "x-user:newUser" -X POST -d '{"tx":"getAssignmentResults","args":["unused"]}' http://localhost:3000/api/cc/invoke

# Test query with newUser not registered with CA;  expect error

curl -H "Content-Type: application/json" -H "x-user:newUser" -X POST -d '{"tx":"getAssignmentResults","args":["unused"]}' http://localhost:3000/api/cc/query


# Test invoke with a function name which does not exist;  expect error

curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"functionDoesNotExist","args":["unused"]}' http://localhost:3000/api/cc/invoke

# Test query with a function name which does not exist;  expect error

curl -H "Content-Type: application/json" -H "x-user:Auctioneer1" -X POST -d '{"tx":"functionDoesNotExist","args":["unused"]}' http://localhost:3000/api/cc/query
