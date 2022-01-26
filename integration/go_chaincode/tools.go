//go:build tools
// +build tools

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package go_chaincode

import (
	_ "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/weaver/relay/fabric"
	_ "github.com/hyperledger-labs/fabric-smart-client/platform/view/services/comm"
	_ "github.com/hyperledger/fabric/cmd/orderer"
	_ "github.com/hyperledger/fabric/cmd/peer"
	_ "github.com/hyperledger/fabric/common/ledger/util/leveldbhelper"
	_ "github.com/hyperledger/fabric/common/metrics/prometheus"
	_ "github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/statecouchdb"
	_ "github.com/hyperledger/fabric/core/ledger/pvtdatastorage"
	_ "github.com/hyperledger/fabric/core/operations"
	_ "github.com/libp2p/go-libp2p-core/network"
)
