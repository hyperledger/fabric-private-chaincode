/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fpcclientsdk enables Go developers to build solutions that interact with FPC chaincode for Hyperledger Fabric.
// The FPC Client SDK builds on top of the Fabric Client SDK Go (https://godoc.org/github.com/hyperledger/fabric-sdk-go)
// and provides FPC specific functionality such as enclave initialization and secure interaction with a FPC chaincode.
// The main goal is to ease the interaction with a FPC chaincode and provide similar experience as offered by normal
// chaincode interaction.
//
// Packages for end developer usage
//
// pkg/client/resmgmt: Provides resource management capabilities such as installing FPC chaincode.
// Reference: https://godoc.org/github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/client/resmgmt
//
// pkg/gateway: Enables Go developers to build client applications using the Hyperledger
// Fabric programming model.
// Reference: https://godoc.org/github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway
//
// Usage samples
//
// $FPC_PATH/samples/application: Illustrates the use of the FPC Client SDK.
// The sample applications can be used with our test-network `$FPC_PATH/samples/deployment/test-network`.
//
package fpcclientsdk
