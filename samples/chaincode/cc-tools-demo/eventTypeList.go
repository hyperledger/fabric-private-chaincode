package main

import (
	"github.com/hyperledger-labs/cc-tools/events"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/eventtypes"
)

var eventTypeList = []events.Event{
	eventtypes.CreateLibraryLog,
}
