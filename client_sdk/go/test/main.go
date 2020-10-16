package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-protos-go/msp"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric/protoutil"
)

func main() {

	os.Setenv("GRPC_TRACE", "all")
	os.Setenv("GRPC_VERBOSITY", "DEBUG")
	os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "INFO")

	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract("ercc")

	log.Println("--> QueryListEnclaveCredentials: ")
	result, err := contract.EvaluateTransaction("QueryListEnclaveCredentials", "myChaincode")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   "Org1MSP",
		IdBytes: []byte("some bytes"),
	}
	serializedUser := protoutil.MarshalOrPanic(serializedIdentity)

	credentials := &protos.Credentials{
		Evidence: []byte("{\"attestation_type\":\"simulated\",\"evidence\":\"MA==\"}"),
		SerializedAttestedData: &any.Any{
			TypeUrl: proto.MessageName(&protos.Attested_Data{}),
			Value: protoutil.MarshalOrPanic(&protos.Attested_Data{
				EnclaveVk: []byte("enclaveVKString"),
				CcParams: &protos.CC_Parameters{
					ChaincodeId: "ercc",
					Version:     "1.0",
					ChannelId:   "mychannel",
					Sequence:    1,
				},
				HostParams: &protos.Host_Parameters{
					PeerIdentity: serializedUser,
				},
			}),
		},
	}

	credentialBase64 := base64.StdEncoding.EncodeToString(protoutil.MarshalOrPanic(credentials))

	log.Println("--> RegisterEnclave: ")
	result, err = contract.SubmitTransaction("RegisterEnclave", credentialBase64)
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	log.Println("============ application-golang ends ============")
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}
