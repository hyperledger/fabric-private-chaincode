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

// ConfidentialEscrowCC implements the chaincode interface
type ConfidentialEscrowCC struct{}

// Init is called during chaincode instantiation
func (t *ConfidentialEscrowCC) Init(stub shim.ChaincodeStubInterface) (response pb.Response) {
	log.Println("ConfidentialEscrowCC: Init called")

	res := InitFunc(stub)
	startupCheckExecuted = true
	if res.Status != 200 {
		return res
	}

	return shim.Success(nil)
}

// InitFunc performs startup checks
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

// Invoke is called for each transaction
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

// logTx logs transaction details
func logTx(stub shim.ChaincodeStubInterface, beginTime time.Time, response *pb.Response) {
	fn, _ := stub.GetFunctionAndParameters()
	log.Printf("%d %s %s %s\n", response.Status, fn, time.Since(beginTime), response.Message)
}
