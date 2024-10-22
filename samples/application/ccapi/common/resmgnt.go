package common

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type fabricResmgtmClient struct {
	Provider context.ClientProvider
	Client   *resmgmt.Client
}
type fabricChannelClient struct {
	Provider context.ChannelProvider
	Client   *channel.Client
}
type fabricLedgerClient struct {
	Provider context.ChannelProvider
	Client   *ledger.Client
}

// Returns a client which has access resource management capabilities
// These are, but not limited to: create channel, query cfg, cc lifecycle...
//
// Function works like this:
//  1. Get sdk
//  2. Use sdk to create a ClientProvider ()
//  3. From client provider create resmgmt Client
// You can then use this .Client to call for specific functionalities
func NewFabricResmgmtClient(orgName, userName string, opts ...resmgmt.ClientOption) (*fabricResmgtmClient, error) {
	sdk, err := GetSDK()
	if err != nil {
		return nil, err
	}

	// Create ClientProvider
	clientProvider := sdk.CreateClientContext(fabsdk.WithOrg(orgName), fabsdk.WithUser(userName))

	// Resource management client is responsible for managing channels (create/update channel)
	// Supply user that has privileges to create channel
	resMgmtClient, err := resmgmt.New(clientProvider, opts...)
	if err != nil {
		return nil, err
	}

	return &fabricResmgtmClient{
		Provider: clientProvider,
		Client:   resMgmtClient,
	}, nil
}

// Returns a client which has channel transaction capabilities
// These are, but not limited to: Execute, Query, Invoke cc...
//
// Function works like this:
//  1. Get sdk
//  2. Use sdk to create a ChannelProvider ()
//  3. From channel provider create channel Client
// You can then use this .Client to call for specific functionalities
func NewFabricChClient(channelName, userName, orgName string) (*fabricChannelClient, error) {
	sdk, err := GetSDK()
	if err != nil {
		return nil, err
	}

	// Create Channel Provider
	chProvider := sdk.CreateChannelContext(channelName, fabsdk.WithUser(userName), fabsdk.WithOrg(orgName))

	// Create Channel's chClient
	chClient, err := channel.New(chProvider)
	if err != nil {
		return nil, err
	}

	return &fabricChannelClient{
		Provider: chProvider,
		Client:   chClient,
	}, nil
}

// Returns a client which can query a channel's underlying ledger,
// such as QueryBlock and QueryConfig
//
// Function works like this:
//  1. Get sdk
//  2. Use sdk to create a ChannelProvider ()
//  3. From channel provider create ledger Client
// You can then use this .Client to call for specific functionalities
func NewFabricLedgerClient(channelName, user, orgName string) (*fabricLedgerClient, error) {
	sdk, err := GetSDK()
	if err != nil {
		return nil, err
	}

	// Create Channel Provider
	chProvider := sdk.CreateChannelContext(channelName, fabsdk.WithUser(user), fabsdk.WithOrg(orgName))
	// chProvider := sdk.CreateChannelContext(channelName, )
	// Create Channel's chClient
	ledgerClient, err := ledger.New(chProvider)
	if err != nil {
		return nil, err
	}

	return &fabricLedgerClient{
		Provider: chProvider,
		Client:   ledgerClient,
	}, nil
}
