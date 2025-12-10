package chaincode

import (
	"log"
	"time"

	"github.com/hyperledger-labs/cc-tools/assets"
	tx "github.com/hyperledger-labs/cc-tools/transactions"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

var startupCheckExecuted = false

// ConfidentialEscrowCC implements the Hyperledger Fabric chaincode interface
// for confidential escrow operations.
type ConfidentialEscrowCC struct{}

// Init is called during chaincode instantiation to initialize the ledger state.
// This method performs startup validation checks for assets and transactions
// to ensure the chaincode is properly configured before accepting invocations.
//
// Parameters:
//
//	stub: The chaincode stub for ledger interaction
//
// Returns:
//
//	response: Peer response indicating success or failure of initialization
func (t *ConfidentialEscrowCC) Init(stub shim.ChaincodeStubInterface) (response pb.Response) {
	log.Println("ConfidentialEscrowCC: Init called")

	res := InitFunc(stub)
	startupCheckExecuted = true
	if res.Status != 200 {
		return res
	}

	return shim.Success(nil)
}

// InitFunc performs comprehensive startup validation checks.
// It verifies that all asset types and transaction definitions are properly
// configured and consistent with the chaincode specification.
//
// Parameters:
//
//	stub: The chaincode stub for ledger interaction
//
// Returns:
//
//	response: Peer response with status 200 on success, error response otherwise
func InitFunc(stub shim.ChaincodeStubInterface) (response pb.Response) {
	defer logTx(stub, time.Now(), &response)

	// Run cc-tools startup checks
	err := assets.StartupCheck()
	if err != nil {
		response = err.GetErrorResponse()
		return
	}

	err = tx.StartupCheck()
	if err != nil {
		response = err.GetErrorResponse()
		return
	}

	log.Println("Confidential Escrow chaincode initialized successfully")
	return shim.Success(nil)
}

// Invoke is called for each transaction invocation on the chaincode.
// It ensures proper initialization has occurred and delegates transaction
// execution to the cc-tools transaction runner.
//
// Parameters:
//
//	stub: The chaincode stub for ledger interaction
//
// Returns:
//
//	response: Peer response containing transaction result or error detail
func (t *ConfidentialEscrowCC) Invoke(stub shim.ChaincodeStubInterface) (response pb.Response) {
	defer logTx(stub, time.Now(), &response)

	// Ensure startup check is executed
	if !startupCheckExecuted {
		log.Println("Running startup check...")
		res := InitFunc(stub)
		if res.Status != 200 {
			return res
		}
		startupCheckExecuted = true
	}

	// Use cc-tools transaction runner
	result, err := tx.Run(stub)
	if err != nil {
		response = err.GetErrorResponse()
		return
	}

	return shim.Success([]byte(result))
}

// logTx logs transaction execution details including status, duration, and any error messages.
// This function is deferred to ensure logging occurs regardless of transaction outcome.
//
// Parameters:
//
//	stub: The chaincode stub for accessing transaction context
//	beginTime: Transaction start time for duration calculation
//	response: Pointer to the peer response for status logging
func logTx(stub shim.ChaincodeStubInterface, beginTime time.Time, response *pb.Response) {
	fn, _ := stub.GetFunctionAndParameters()
	log.Printf("%d %s %s %s\n", response.Status, fn, time.Since(beginTime), response.Message)
}
