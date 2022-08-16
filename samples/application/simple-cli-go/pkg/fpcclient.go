/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkg

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	cfg "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/pkg/errors"
)

type Client struct {
	contract fpc.Contract
}

func NewClient(config *Config) *Client {
	return &Client{contract: newContract(config)}
}

func findSigningCert(mspConfigPath string) (string, error) {
	p := filepath.Join(mspConfigPath, "signcerts")
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return "", errors.Wrapf(err, "error while searching pem in %s", mspConfigPath)
	}

	// return first pem we find
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pem") {
			return filepath.Join(p, f.Name()), nil
		}
	}

	return "", errors.Errorf("cannot find pem in %s", mspConfigPath)
}

func populateWallet(wallet *gateway.Wallet, config *Config) error {
	logger.Debugf("============ Populating wallet ============")
	certPath, err := findSigningCert(config.CorePeerMSPConfigPath)
	if err != nil {
		return err
	}

	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(config.CorePeerMSPConfigPath, "keystore")
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

	identity := gateway.NewX509Identity(config.CorePeerLocalMSPID, string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func newContract(config *Config) fpc.Contract {

	wallet := gateway.NewInMemoryWallet()
	err := populateWallet(wallet, config)
	if err != nil {
		logger.Fatalf("Failed to populate wallet contents: %v", err)
	}

	gw, err := gateway.Connect(
		gateway.WithConfig(cfg.FromFile(filepath.Clean(config.GatewayConfigPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		logger.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork(config.ChannelId)
	if err != nil {
		logger.Fatalf("Failed to get network: %v", err)
	}

	// Get FPC Contract
	contract := fpc.GetContract(network, config.ChaincodeId)
	return contract
}

func (c *Client) Invoke(function string, args ...string) string {
	logger.Debugf("--> Invoke FPC chaincode with %s %s", function, args)
	result, err := c.contract.SubmitTransaction(function, args...)
	if err != nil {
		logger.Fatalf("Failed to Submit transaction: %v", err)
	}
	logger.Debugf("--> Result: %s", string(result))
	return string(result)
}

func (c *Client) Query(function string, args ...string) string {
	logger.Debugf("--> Query FPC chaincode with %s %s", function, args)
	result, err := c.contract.EvaluateTransaction(function, args...)
	if err != nil {
		logger.Fatalf("Failed to evaluate transaction: %v", err)
	}
	logger.Debugf("--> Result: %s", string(result))
	return string(result)
}
