package common

import (
	"fmt"
	"log"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type sdk struct {
	// sdk belongs to org defined in the configsdk.yaml file
	Sdk  *fabsdk.FabricSDK
	Path string
}

// CreateContext allows creation of transactions using the supplied identity as the credential.
func (s *sdk) CreateClientContext(options ...fabsdk.ContextOption) context.ClientProvider {
	return s.Sdk.Context(options...)
}

func (s *sdk) CreateChannelContext(channelName string, options ...fabsdk.ContextOption) context.ChannelProvider {
	return s.Sdk.ChannelContext(channelName, options...)
}

// Log config path which sdk was created
func (s *sdk) LogPath() {
	log.Printf("sdk created from '%s'", s.Path)
}

// Singleton sdk instance
var instance *sdk

// GetSDK returns a fabric sdk instance.
//
//	A new sdk is created if:
//	- it is the first time it is beeing used, or
//	- new sdk options are given
//
// Otherwise, it returns the one previoulsy created.
// If options are given, the new sdk is not a singleton, and must
// be closed by whoever invoked it.
//
// The configsdk file can be set via environment variable and defaults
// to './config/configsdk.yaml'
func GetSDK(sdkOpts ...fabsdk.Option) (*sdk, error) {

	// return new sdk instance if sdkOpts are given.
	// user must close sdk
	if len(sdkOpts) != 0 {
		cfgPath := getCfgPath()
		configOpt := config.FromFile(cfgPath)
		s, err := fabsdk.New(configOpt, sdkOpts...)

		return &sdk{
			Sdk:  s,
			Path: cfgPath,
		}, err
	}

	if instance == nil {
		cfgPath := getCfgPath()
		configOpt := config.FromFile(cfgPath)
		s, err := fabsdk.New(configOpt)
		if err != nil {
			return nil, err
		}

		instance = &sdk{
			Sdk:  s,
			Path: cfgPath,
		}
		instance.LogPath()
	}

	return instance, nil
}

// getCfgPath parses path for the configsdk
// from environmet, and defaults to './config/configsdk.yaml'
func getCfgPath() (cfgPath string) {
	cfgPath = os.Getenv("SDK_PATH")
	if cfgPath == "" {
		cfgPath = "./config/configsdk.yaml"
	}
	return
}

// GetClientOrg returns the name of the client organization
func GetClientOrg() string {
	sdk, err := GetSDK()
	if err != nil {
		return ""
	}

	cfg, err := sdk.Sdk.Config()
	if err != nil {
		return ""
	}

	i, ok := cfg.Lookup("client")
	if !ok {
		return ""
	}
	m, ok := i.(map[string]interface{})
	if !ok {
		return ""
	}

	org := m["organization"]
	orgName, ok := org.(string)
	if !ok {
		return ""
	}

	return orgName
}

func GetCryptoPath() string {
	sdk, err := GetSDK()
	if err != nil {
		return ""
	}

	cfg, err := sdk.Sdk.Config()
	if err != nil {
		return ""
	}

	i, ok := cfg.Lookup("client.cryptoconfig.path")
	if !ok {
		return ""
	}
	basePath, _ := i.(string)

	i, ok = cfg.Lookup(fmt.Sprintf("organizations.%s.cryptoPath", os.Getenv("ORG")))
	if !ok {
		return ""
	}

	certPath, _ := i.(string)
	return basePath + "/" + certPath
}

func GetTLSCACert() string {
	sdk, err := GetSDK()
	if err != nil {
		return ""
	}

	cfg, err := sdk.Sdk.Config()
	if err != nil {
		return ""
	}

	i, ok := cfg.Lookup("client.tlsCerts.client.cacertfile")
	if !ok {
		return ""
	}

	certPath, _ := i.(string)
	return certPath
}

func GetMSPID() string {
	sdk, err := GetSDK()
	if err != nil {
		return ""
	}

	cfg, err := sdk.Sdk.Config()
	if err != nil {
		return ""
	}

	i, ok := cfg.Lookup(fmt.Sprintf("organizations.%s.mspid", os.Getenv("ORG")))
	if !ok {
		return ""
	}

	mspid, _ := i.(string)
	return mspid
}

// Closes sdk instance if it was created
func CloseSDK() {
	if instance != nil {
		instance.Sdk.Close()
		instance = nil
	}
}
