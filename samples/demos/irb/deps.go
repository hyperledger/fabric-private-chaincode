//go:build deps

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package irb

import (
	fscnode "github.com/hyperledger-labs/fabric-smart-client/node"
	viewregistry "github.com/hyperledger-labs/fabric-smart-client/platform/view/services/view"
)

// note that this file keeps the fsc runtime deps into the go.mod file;
// we do this to avoid errors during test builds via fsc; this is clearly a hack - don't try at home!

func main() {
	node := fscnode.New()
	node.Execute(func() error {
		registry := viewregistry.GetRegistry(n)
		_ = registry
		return nil
	})
}
