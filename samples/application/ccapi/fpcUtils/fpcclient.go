/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpcUtils

import (
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
)

type Client struct {
	contract fpc.Contract
}

func NewClient(config *Config) *Client {
	return &Client{contract: newContract(config)}
}

func newContract(config *Config) fpc.Contract {
	gw, _, sign, err := connectGateway(config)
	if err != nil {
		logger.Fatalf("Failed to connect to gateway: %v", err)
	}

	network := gw.GetNetwork(config.ChannelId)

	contract := fpc.GetContract(network, config.ChaincodeId, sign, peerDialer(config))
	return contract
}

func (c *Client) Invoke(function string, args ...string) string {
	logger.Infof("--> Invoke FPC chaincode with %s %s", function, args)
	result, err := c.contract.SubmitTransaction(function, args...)
	if err != nil {
		logger.Infof("Failed to Submit transaction: %v", err)
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
