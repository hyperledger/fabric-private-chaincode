// The internal package contains some helper interfaces and adapters to remove the direct dependency to Fabric Go SDK components.
package internal

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// Network interface that is needed by the FPC contract implementation
type Network interface {
	GetContract(chaincodeID string) *gateway.Contract
}

// Transaction interface that is needed by the FPC contract implementation
type Transaction interface {
	Evaluate(args ...string) ([]byte, error)
}

// Contract interface
type Contract interface {
	Name() string
	EvaluateTransaction(name string, args ...string) ([]byte, error)
	SubmitTransaction(name string, args ...string) ([]byte, error)
	CreateTransaction(name string, opts ...gateway.TransactionOption) (Transaction, error)
	RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error)
	Unregister(registration fab.Registration)
}

// ContractAdapter wraps a gateway.Contract with the Contract interface
type ContractAdapter struct {
	Contract *gateway.Contract
}

func (f *ContractAdapter) Name() string {
	return f.Contract.Name()
}

func (f *ContractAdapter) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	return f.Contract.EvaluateTransaction(name, args...)
}

func (f *ContractAdapter) SubmitTransaction(name string, args ...string) ([]byte, error) {
	return f.Contract.SubmitTransaction(name, args...)
}

func (f *ContractAdapter) CreateTransaction(name string, opts ...gateway.TransactionOption) (Transaction, error) {
	return f.Contract.CreateTransaction(name, opts...)
}

func (f *ContractAdapter) RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	return f.Contract.RegisterEvent(eventFilter)
}

func (f *ContractAdapter) Unregister(registration fab.Registration) {
	f.Contract.Unregister(registration)
}
