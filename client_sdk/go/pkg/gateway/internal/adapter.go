// The internal package contains some helper interfaces and adapters to remove the direct dependency to Fabric Go SDK components.
package internal

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// Network interface that is needed by the FPC contract implementation
type Network interface {
	GetContract(chaincodeID string) *gateway.Contract
}
