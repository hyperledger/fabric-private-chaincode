package chaincode

import (
	"os"

	"github.com/hyperledger-labs/ccapi/common"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/pkg/errors"
)

func InvokeGateway(channelName, chaincodeName, txName, user string, args []string, transientArgs []byte, endorsingOrgs []string) ([]byte, error) {
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

	// Make transient request
	transientMap := make(map[string][]byte)
	transientMap["@request"] = transientArgs

	// Invoke transaction
	if transientArgs != nil && len(endorsingOrgs) > 0 {
		return contract.Submit(txName,
			client.WithArguments(args...),
			client.WithTransient(transientMap),
			client.WithEndorsingOrganizations(endorsingOrgs...),
		)
	}

	if transientArgs != nil {
		return contract.Submit(txName,
			client.WithArguments(args...),
			client.WithTransient(transientMap),
		)
	}

	if len(endorsingOrgs) > 0 {
		return contract.Submit(txName,
			client.WithArguments(args...),
			client.WithEndorsingOrganizations(endorsingOrgs...),
		)
	}

	return contract.SubmitTransaction(txName, args...)
}
