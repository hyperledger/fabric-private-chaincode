/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkg

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/hyperledger/fabric/common/flogging"
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

	return e
}

type extendedFlogger struct {
	*flogging.FabricLogger
}

func (e *extendedFlogger) Fatalln(v ...interface{}) {
	e.Fatal(v...)
}

func (e *extendedFlogger) Panicln(v ...interface{}) {
	e.Panic(v...)
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
	e.Debug(args...)
}

func (e *extendedFlogger) Infoln(args ...interface{}) {
	e.Info(args...)
}

func (e *extendedFlogger) Warnln(args ...interface{}) {
	e.Warn(args...)
}

func (e *extendedFlogger) Errorln(args ...interface{}) {
	e.Error(args...)
}
