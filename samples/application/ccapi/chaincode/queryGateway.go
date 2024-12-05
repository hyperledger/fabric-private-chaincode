package chaincode

import (
	"os"

	"github.com/hyperledger-labs/ccapi/common"
	"github.com/pkg/errors"
)

func QueryGateway(channelName, chaincodeName, txName, user string, args []string) ([]byte, error) {
	// Gateway endpoint
	endpoint := os.Getenv("FABRIC_GATEWAY_ENDPOINT")

	// Create client grpc connection
	grpcConn, err := common.CreateGrpcConnection(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create grpc connection")
	}
	defer grpcConn.Close()

	// Create gateway connection
	gw, err := common.CreateGatewayConnection(grpcConn, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gateway connection")
	}
	defer gw.Close()

	// Obtain smart contract deployed on the network.
	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	// Query transaction
	if len(args) == 0 {
		return contract.EvaluateTransaction(txName)
	}

	return contract.EvaluateTransaction(txName, args...)
}
