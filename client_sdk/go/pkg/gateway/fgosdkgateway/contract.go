package fgosdkgateway

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"

	gateway2 "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
)

// Network interface that is needed by the FPC contract implementation
type Network interface {
	GetContract(chaincodeID string) *gateway.Contract
}

type contract struct {
	c *gateway.Contract
}

func (c *contract) Name() string {
	return c.c.Name()
}

func (c *contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	return c.c.EvaluateTransaction(name, args...)
}

func (c *contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	return c.c.SubmitTransaction(name, args...)
}

func (c *contract) CreateTransaction(name string, peerEndpoints ...string) (gateway2.Transaction, error) {
	return c.c.CreateTransaction(name, gateway.WithEndorsingPeers(peerEndpoints...))
}

type contractProvider struct {
	network Network
}

func (c *contractProvider) GetContract(id string) gateway2.Contract {
	return &contract{c.network.GetContract(id)}
}

func GetContract(network Network, chaincodeID string) *gateway2.ContractState {
	return gateway2.GetContract(
		&contractProvider{network: network}, chaincodeID,
	)
}

