/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpcUtils

import (
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"google.golang.org/grpc"
)

type Client struct {
	gw       *client.Gateway
	conn     *grpc.ClientConn
	contract fpc.Contract
}

func NewClient(config *Config) (*Client, error) {
	gw, conn, sign, err := connectGateway(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	network := gw.GetNetwork(config.ChannelId)
	contract := fpc.GetContract(network, config.ChaincodeId, sign, peerDialer(config))
	return &Client{gw: gw, conn: conn, contract: contract}, nil
}

func (c *Client) Close() {
	c.gw.Close()
	c.conn.Close()
}

func (c *Client) Invoke(function string, args ...string) (string, error) {
	logger.Infof("--> Invoke FPC chaincode with %s %s", function, args)
	result, err := c.contract.SubmitTransaction(function, args...)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}
	logger.Debugf("--> Result: %s", string(result))
	return string(result), nil
}

func (c *Client) Query(function string, args ...string) (string, error) {
	logger.Debugf("--> Query FPC chaincode with %s %s", function, args)
	result, err := c.contract.EvaluateTransaction(function, args...)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate transaction: %w", err)
	}
	logger.Debugf("--> Result: %s", string(result))
	return string(result), nil
}
