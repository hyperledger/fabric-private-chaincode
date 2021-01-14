package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

var testNetworkPath string

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		testNetworkPath,
		"config",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "peer.pem")
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

	identity := gateway.NewX509Identity("SampleOrg", string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func setupNetwork(channel string) (*gateway.Network, error) {
	os.Setenv("GRPC_TRACE", "all")
	os.Setenv("GRPC_VERBOSITY", "DEBUG")
	os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "INFO")

	fpcPath := os.Getenv("FPC_PATH")
	if fpcPath == "" {
		return nil, fmt.Errorf("FPC_PATH not set")
	}
	testNetworkPath = filepath.Join(fpcPath, "integration")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "false")
	if err != nil {
		return nil, fmt.Errorf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		return nil, fmt.Errorf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			return nil, fmt.Errorf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		testNetworkPath,
		"config",
		"msp",
		"connection.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channel)
	if err != nil {
		return nil, fmt.Errorf("Failed to get network: %v", err)
	}

	return network, nil
}
