package eventtypes

import "github.com/hyperledger-labs/cc-tools/events"

// CreateLibraryLog is a log to be emitted on the CCAPI when a library is created
var CreateLibraryLog = events.Event{
	Tag:         "createLibraryLog",
	Label:       "Create Library Log",
	Description: "Log of a library creation",
	Type:        events.EventLog,                  // Event funciton is to log on the CCAPI
	BaseLog:     "New library created",            // BaseLog is a base message to be logged
	Receivers:   []string{"$org2MSP", "$Org1MSP"}, // Receivers are the MSPs that will receive the event
}
