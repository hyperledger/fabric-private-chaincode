/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpcUtils

import (
	"fmt"

	"github.com/hyperledger/fabric-lib-go/common/flogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
)

var logger = flogging.MustGetLogger("fpc.cli")

func init() {
	logging.Initialize(&provider{})
}

type provider struct {
}

func (p *provider) GetLogger(module string) api.Logger {
	name := "client.sdk-go"
	e := &extendedFlogger{flogging.MustGetLogger(name)}

	return e.FabricLogger
}

type extendedFlogger struct {
	*flogging.FabricLogger
}

func (e *extendedFlogger) Fatalln(v ...interface{}) {
	e.Fatalln(v...)
}

func (e *extendedFlogger) Panicln(v ...interface{}) {
	e.Panicln(v...)
}

func (e *extendedFlogger) Print(v ...interface{}) {
	fmt.Print(v...)
}

func (e *extendedFlogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (e *extendedFlogger) Println(v ...interface{}) {
	fmt.Println(v...)

}

func (e *extendedFlogger) Debugln(args ...interface{}) {
	e.Debugln(args...)
}

func (e *extendedFlogger) Infoln(args ...interface{}) {
	e.Infoln(args...)
}

func (e *extendedFlogger) Warnln(args ...interface{}) {
	e.Warnln(args...)
}

func (e *extendedFlogger) Errorln(args ...interface{}) {
	e.Errorln(args...)
}
