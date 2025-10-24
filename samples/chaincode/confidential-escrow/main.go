package main

import (
	"flag"
	"log"

	cc "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode"
)

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
