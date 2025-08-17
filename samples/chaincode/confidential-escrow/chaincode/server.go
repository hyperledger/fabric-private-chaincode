package chaincode

import (
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
)

// RunCCaaS starts the chaincode as a service
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

// StartChaincode starts the chaincode in appropriate mode
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
