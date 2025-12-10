package chaincode

import (
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
)

// RunCCaaS starts the chaincode as a service using the Chaincode as a Service (CCaaS) model.
// It reads server configuration from environment variables and optionally wraps the chaincode
// with FPC (Fabric Private Chaincode) for confidential execution.
//
// Environment Variables:
//
//	CHAINCODE_SERVER_ADDRESS: The address and port for the chaincode server
//	CHAINCODE_PKG_ID: The chaincode package identifier
//	FPC_ENABLED: Set to "true" to enable FPC confidential execution
//
// Returns:
//
//	error: Error if server fails to start, nil on successful startup
func RunCCaaS() error {
	address := os.Getenv("CHAINCODE_SERVER_ADDRESS")
	ccid := os.Getenv("CHAINCODE_PKG_ID") // FPC uses PKG_ID

	var cc shim.Chaincode
	if os.Getenv("FPC_ENABLED") == "true" {
		cc = fpc.NewPrivateChaincode(new(ConfidentialEscrowCC))
	} else {
		cc = new(ConfidentialEscrowCC)
	}

	server := &shim.ChaincodeServer{
		CCID:    ccid,
		Address: address,
		CC:      cc,
		TLSProps: shim.TLSProperties{
			Disabled: true, // TLS handled by FPC
		},
	}

	return server.Start()
}

// StartChaincode initializes and starts the chaincode in the appropriate execution mode.
// It determines the startup mode based on environment variables and handles both
// CCaaS (Chaincode as a Service) and direct execution modes, with optional FPC support.
//
// Environment Variables:
//
//	RUN_CCAAS: Set to "true" to run in CCaaS mode
//	FPC_ENABLED: Set to "true" to enable FPC confidential execution
//
// Returns:
//
//	error: Error if chaincode fails to start, nil on successful startup
func StartChaincode() error {
	if os.Getenv("RUN_CCAAS") == "true" {
		return RunCCaaS()
	} else {
		// Fallback for direct start
		if os.Getenv("FPC_ENABLED") == "true" {
			return shim.Start(fpc.NewPrivateChaincode(new(ConfidentialEscrowCC)))
		} else {
			return shim.Start(new(ConfidentialEscrowCC))
		}
	}
}
