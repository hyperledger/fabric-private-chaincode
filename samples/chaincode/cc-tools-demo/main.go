package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/events"
	sw "github.com/hyperledger-labs/cc-tools/stubwrapper"
	tx "github.com/hyperledger-labs/cc-tools/transactions"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/assettypes"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/datatypes"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/header"

	fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

var startupCheckExecuted = false

func SetupCC() error {
	tx.InitHeader(tx.Header{
		Name:    header.Name,
		Version: header.Version,
		Colors:  header.Colors,
		Title:   header.Title,
	})

	assets.InitDynamicAssetTypeConfig(assettypes.DynamicAssetTypes)

	tx.InitTxList(txList)

	err := assets.CustomDataTypes(datatypes.CustomDataTypes)
	if err != nil {
		fmt.Printf("Error injecting custom data types: %s", err)
		return err
	}
	assets.InitAssetList(append(assetTypeList, assettypes.CustomAssets...))

	events.InitEventList(eventTypeList)

	return nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	// Generate collection json
	genFlag := flag.Bool("g", false, "Enable collection generation")
	flag.Bool("orgs", false, "List of orgs to generate collection for")
	flag.Parse()
	if *genFlag {
		listOrgs := flag.Args()
		generateCollection(listOrgs)
		return
	}

	log.Printf("Starting chaincode %s version %s\n", header.Name, header.Version)

	err := SetupCC()
	if err != nil {
		return
	}

	if os.Getenv("RUN_CCAAS") == "true" {
		err = runCCaaS()
	} else {
		if os.Getenv("FPC_ENABLED") == "true" {
			err = shim.Start(fpc.NewPrivateChaincode(new(CCDemo)))
		} else {
			err = shim.Start(new(CCDemo))
		}
	}

	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}

func runCCaaS() error {
	address := os.Getenv("CHAINCODE_SERVER_ADDRESS")
	ccid := os.Getenv("CHAINCODE_PKG_ID")

	tlsProps, err := getTLSProperties()
	if err != nil {
		return err
	}

	var cc shim.Chaincode

	if os.Getenv("FPC_ENABLED") == "true" {
		cc = fpc.NewPrivateChaincode(new(CCDemo))
	} else {
		cc = new(CCDemo)
	}

	server := &shim.ChaincodeServer{
		CCID:     ccid,
		Address:  address,
		CC:       cc,
		TLSProps: *tlsProps,
	}

	return server.Start()
}

func getTLSProperties() (*shim.TLSProperties, error) {
	if enableTLS := os.Getenv("TLS_ENABLED"); enableTLS != "true" {
		return &shim.TLSProperties{
			Disabled: true,
		}, nil
	}

	log.Printf("TLS enabled")

	// Get key
	keyPath := os.Getenv("KEY_PATH")
	key, err := os.ReadFile(keyPath)

	if err != nil {
		fmt.Println("Failed to read key file")
		return nil, err
	}

	// Get cert
	certPath := os.Getenv("CERT_PATH")
	cert, err := os.ReadFile(certPath)
	if err != nil {
		fmt.Println("Failed to read cert file")
		return nil, err
	}

	// Get CA cert
	clientCertPath := os.Getenv("CA_CERT_PATH")
	caCert, err := os.ReadFile(clientCertPath)
	if err != nil {
		fmt.Println("Failed to read CA cert file")
		return nil, err
	}

	return &shim.TLSProperties{
		Disabled:      false,
		Key:           key,
		Cert:          cert,
		ClientCACerts: caCert,
	}, nil
}

// CCDemo implements the shim.Chaincode interface
type CCDemo struct{}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *CCDemo) Init(stub shim.ChaincodeStubInterface) (response pb.Response) {

	res := InitFunc(stub)
	startupCheckExecuted = true
	if res.Status != 200 {
		return res
	}

	response = shim.Success(nil)
	return
}

func InitFunc(stub shim.ChaincodeStubInterface) (response pb.Response) {
	// Defer logging function
	defer logTx(stub, time.Now(), &response)

	if assettypes.DynamicAssetTypes.Enabled {
		sw := &sw.StubWrapper{
			Stub: stub,
		}
		err := assets.RestoreAssetList(sw, true)
		if err != nil {
			response = err.GetErrorResponse()
			return
		}
	}

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

	response = shim.Success(nil)
	return
}

// Invoke is called per transaction on the chaincode.
func (t *CCDemo) Invoke(stub shim.ChaincodeStubInterface) (response pb.Response) {
	// Defer logging function
	defer logTx(stub, time.Now(), &response)
	myId, _ := cid.GetMSPID(stub)
	fmt.Println("org is.......", myId)

	if !startupCheckExecuted {
		fmt.Println("Running startup check...")
		res := InitFunc(stub)
		if res.Status != 200 {
			return res
		}
		startupCheckExecuted = true
	}

	var result []byte

	result, err := tx.Run(stub)

	if err != nil {
		response = err.GetErrorResponse()
		return
	}
	response = shim.Success([]byte(result))
	return
}

func logTx(stub shim.ChaincodeStubInterface, beginTime time.Time, response *pb.Response) {
	fn, _ := stub.GetFunctionAndParameters()
	log.Printf("%d %s %s %s\n", response.Status, fn, time.Since(beginTime), response.Message)
}
