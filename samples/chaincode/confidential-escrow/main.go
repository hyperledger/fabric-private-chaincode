package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/events"
	tx "github.com/hyperledger-labs/cc-tools/transactions"

	asset "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/assets"
	header "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/header"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

var startupCheckExecuted = false

// PPS:REMOVE
var (
	txList        = []tx.Transaction{} // Empty for now
	assetTypeList = []assets.AssetType{
		asset.Wallet,
		asset.DigitalAssetToken,
		asset.UserDirectory,
	}
	eventTypeList = []events.Event{} // Empty for now
)

// SetupCC initializes the chaincode with assets and transactions
func SetupCC() error {
	// Initialize header info
	tx.InitHeader(tx.Header{
		Name:    header.Name,
		Version: header.Version,
		Colors:  header.Colors,
		Title:   header.Title,
	})

	// Initialize transaction and asset lists (we'll create these next)
	tx.InitTxList(txList)
	assets.InitAssetList(assetTypeList)
	events.InitEventList(eventTypeList)

	return nil
}

// PPS:REMOVE
func generateCollection(orgs []string) {
	fmt.Println("Collection generation called with orgs:", orgs)
	fmt.Println("Collection generation not implemented yet")
	// Exit after generating (like in cc-tools-demo)
	os.Exit(0)
}

// main function starts the chaincode
func main() {
	// Handle collection generation flag (cc-tools feature)
	genFlag := flag.Bool("g", false, "Enable collection generation")
	flag.Bool("orgs", false, "List of orgs to generate collection for")
	flag.Parse()
	if *genFlag {
		listOrgs := flag.Args()
		generateCollection(listOrgs)
		return
	}

	log.Printf("Starting Confidential Escrow Chaincode v1.0.0")

	// Setup CC-Tools components
	err := SetupCC()
	if err != nil {
		log.Printf("Error setting up chaincode: %s", err)
		return
	}

	// Run as Chaincode-as-a-Service (CcaaS) - required for FPC
	if os.Getenv("RUN_CCAAS") == "true" {
		err = runCCaaS()
	} else {
		// Fallback for direct start
		if os.Getenv("FPC_ENABLED") == "true" {
			err = shim.Start(fpc.NewPrivateChaincode(new(ConfidentialEscrowCC)))
		} else {
			err = shim.Start(new(ConfidentialEscrowCC))
		}
	}

	if err != nil {
		log.Printf("Error starting chaincode: %s", err)
	}
}

// runCCaaS starts the chaincode as a service
func runCCaaS() error {
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
