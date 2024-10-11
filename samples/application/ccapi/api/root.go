/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-private-chaincode/samples/application/cc-tools-demo/pkg"
)

var (
	config *pkg.Config
)

func InitConfig() {

	getStrEnv := func(key string) string {
		val := os.Getenv(key)
		if val == "" {
			panic(fmt.Sprintf("%s not set", key))
		}
		return val
	}

	getBoolEnv := func(key string) bool {
		val := getStrEnv(key)
		ret, err := strconv.ParseBool(val)
		if err != nil {
			if val == "" {
				panic(fmt.Sprintf("invalid bool value for %s", key))
			}
		}
		return ret
	}

	config = &pkg.Config{
		CorePeerAddress:         getStrEnv("CORE_PEER_ADDRESS"),
		CorePeerId:              getStrEnv("CORE_PEER_ID"),
		CorePeerLocalMSPID:      getStrEnv("CORE_PEER_LOCALMSPID"),
		CorePeerMSPConfigPath:   getStrEnv("CORE_PEER_MSPCONFIGPATH"),
		CorePeerTLSCertFile:     getStrEnv("CORE_PEER_TLS_CERT_FILE"),
		CorePeerTLSEnabled:      getBoolEnv("CORE_PEER_TLS_ENABLED"),
		CorePeerTLSKeyFile:      getStrEnv("CORE_PEER_TLS_KEY_FILE"),
		CorePeerTLSRootCertFile: getStrEnv("CORE_PEER_TLS_ROOTCERT_FILE"),
		OrdererCA:               getStrEnv("ORDERER_CA"),
		ChaincodeId:             getStrEnv("CC_NAME"),
		ChannelId:               getStrEnv("CHANNEL_NAME"),
		GatewayConfigPath:       getStrEnv("GATEWAY_CONFIG"),
	}
}
