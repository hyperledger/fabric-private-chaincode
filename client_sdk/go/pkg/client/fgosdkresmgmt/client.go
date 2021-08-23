package fgosdkresmgmt

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"

	fpcmgmt "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt"
)

type Client struct {
	*resmgmt.Client
	fpcmgmt *fpcmgmt.Client
}

func NewClient(ctxProvider context.ClientProvider, opts ...resmgmt.ClientOption) (*Client, error) {
	// get resource management client
	c, err := resmgmt.New(ctxProvider, opts...)
	if err != nil {
		return nil, err
	}

	fpcmgmt, err := fpcmgmt.New(NewChannelClientProvider(ctxProvider).ChannelClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:  c,
		fpcmgmt: fpcmgmt,
	}, nil
}

func (c *Client) LifecycleInitEnclave(channelId string, req fpcmgmt.LifecycleInitEnclaveRequest) (string, error) {
	return c.fpcmgmt.LifecycleInitEnclave(channelId, req)
}
