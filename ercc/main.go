package main

import (
	"log"
	"os"

	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/registry"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type serverConfig struct {
	CCID    string
	Address string
}

func main() {
	//ercc, err := contractapi.NewChaincode(&registry.Contract{})
	//if err != nil {
	//	log.Panicf("Error creating registry chaincode: %v", err)
	//}
	//
	//if err := ercc.Start(); err != nil {
	//	log.Panicf("Error starting registry chaincode: %v", err)
	//}

	// See chaincode.env.example
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	c := &registry.Contract{}
	c.Verifier = attestation.NewVerifier()
	c.PEvaluator = utils.NewPolicyEvaluator()

	ercc, err := contractapi.NewChaincode(c)

	if err != nil {
		log.Panicf("error create enclave registry chaincode: %s", err)
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      ercc,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	log.Println("starting enclave registry server")

	if err := server.Start(); err != nil {
		log.Panicf("error starting enclave registry chaincode: %s", err)
	}
}
