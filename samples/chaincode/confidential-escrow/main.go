package main

import (
	"flag"
	"log"

	cc "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode"
)

// main is the entry point for the Confidential Escrow chaincode.
// It handles command-line flags for collection generation and starts the chaincode server.
//
// Flags:
//
//	-g: Enable collection generation mode
//	-orgs: List of organization names for collection configuration
//
// When collection generation is enabled, the application generates
// collection configurations and exits. Otherwise, it initializes
// and starts the chaincode service.
func main() {
	// Handle collection generation flag
	genFlag := flag.Bool("g", false, "Enable collection generation")
	flag.Bool("orgs", false, "List of orgs to generate collection for")
	flag.Parse()
	if *genFlag {
		listOrgs := flag.Args()
		cc.GenerateCollection(listOrgs)
		return
	}

	log.Printf("Starting Confidential Escrow Chaincode v1.0.0")

	// Setup CC-Tools components
	err := cc.SetupCC()
	if err != nil {
		log.Printf("Error setting up chaincode: %s", err)
		return
	}

	// Start chaincode
	err = cc.StartChaincode()
	if err != nil {
		log.Printf("Error starting chaincode: %s", err)
	}
}
