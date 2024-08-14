/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkg

type Config struct {
	CorePeerAddress         string
	CorePeerId              string
	CorePeerLocalMSPID      string
	CorePeerMSPConfigPath   string
	CorePeerTLSCertFile     string
	CorePeerTLSEnabled      bool
	CorePeerTLSKeyFile      string
	CorePeerTLSRootCertFile string
	OrdererCA               string
	FpcPath                 string
	ChaincodeId             string
	ChannelId               string
	GatewayConfigPath       string
}
