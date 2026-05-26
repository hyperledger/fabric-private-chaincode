/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/api"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric"
)

// FPCPlatformFactory wraps FSC's fabric PlatformFactory and registers the FPC NWO extension.
// Use this instead of fabric.NewPlatformFactory() in integration tests and samples that use FPC.
//
// Usage:
//
//	ii, err := integration.Generate(port, false, Topology()...)
//	ii.RegisterPlatformFactory(fpcnwo.NewPlatformFactory())
type FPCPlatformFactory struct {
	inner api.PlatformFactory
}

// NewPlatformFactory returns an FPCPlatformFactory that wraps the standard fabric platform
// factory and automatically wires in the FPC NWO extension.
func NewPlatformFactory() *FPCPlatformFactory {
	return &FPCPlatformFactory{inner: fabric.NewPlatformFactory()}
}

func (f *FPCPlatformFactory) Name() string {
	return f.inner.Name()
}

func (f *FPCPlatformFactory) New(registry api.Context, t api.Topology, builder api.Builder) api.Platform {
	p := f.inner.New(registry, t, builder)
	fp, ok := p.(*fabric.Platform)
	if !ok {
		return p
	}
	fp.Network.AddExtension(NewExtension(fp.Network))
	return fp
}
