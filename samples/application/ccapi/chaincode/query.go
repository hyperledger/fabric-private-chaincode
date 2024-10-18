package chaincode

import (
	"net/http"
	"os"

	"github.com/hyperledger-labs/ccapi/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
)

func Query(channelName, ccName, txName, user string, txArgs [][]byte) (*channel.Response, int, error) {
	// create channel manager
	fabMngr, err := common.NewFabricChClient(channelName, user, os.Getenv("ORG"))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Execute chaincode with channel's client
	rq := channel.Request{ChaincodeID: ccName, Fcn: txName}
	if len(txArgs) > 0 {
		rq.Args = txArgs
	}

	res, err := fabMngr.Client.Query(rq, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		status := extractStatusCode(err.Error())
		return nil, status, err
	}

	return &res, http.StatusOK, nil
}
