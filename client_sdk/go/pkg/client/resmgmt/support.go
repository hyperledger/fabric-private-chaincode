package resmgmt

import (
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
)

type GoSDKChannelClient interface {
	Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	Execute(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	InvokeHandler(handler invoke.Handler, request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	RegisterChaincodeEvent(chainCodeID string, eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error)
	UnregisterChaincodeEvent(registration fab.Registration)
}

type channelClient struct {
	goSDKChannelClient GoSDKChannelClient
}

func (c *channelClient) Query(chaincodeID string, fcn string, args [][]byte, targetEndpoints ...string) ([]byte, error) {
	initRequest := channel.Request{
		ChaincodeID: chaincodeID,
		Fcn:         fcn,
		Args:        args,
	}

	var initOpts []channel.RequestOption
	initOpts = append(initOpts, channel.WithRetry(retry.Opts{Attempts: 0}))
	initOpts = append(initOpts, channel.WithTargetEndpoints(targetEndpoints...))

	// send query to create (init) enclave at the target peer
	initResponse, err := c.goSDKChannelClient.Query(initRequest, initOpts...)
	if err != nil {
		return nil, err
	}
	return initResponse.Payload, nil
}

func (c *channelClient) Execute(chaincodeID string, fcn string, args [][]byte) (string, error) {
	request := channel.Request{
		ChaincodeID: chaincodeID,
		Fcn:         fcn,
		Args:        args,
	}

	var opts []channel.RequestOption
	// TODO translate `resmgmt.RequestOption` to `channel.Option` options so we can pass it to execute
	//opts = append(opts, options...)

	// invoke registerEnclave at enclave registry
	response, err := c.goSDKChannelClient.Execute(request, opts...)
	if err != nil {
		return "", err
	}

	return string(response.TransactionID), nil
}

type ChannelClientProvider struct {
	ctxProvider context.ClientProvider
}

func NewChannelClientProvider(ctxProvider context.ClientProvider) *ChannelClientProvider {
	return &ChannelClientProvider{ctxProvider: ctxProvider}
}

func (c *ChannelClientProvider) ChannelClient(id string) (lifecycle.ChannelClient, error) {
	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(c.ctxProvider, id)
	}
	client, err := channel.New(channelProvider)
	if err != nil {
		return nil, err
	}
	return &channelClient{goSDKChannelClient: client}, nil
}
