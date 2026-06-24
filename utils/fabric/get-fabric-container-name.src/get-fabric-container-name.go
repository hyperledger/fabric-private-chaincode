/*
* Copyright 2019 Intel Corporation
*
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"regexp"
	"strings"
)

var (
	vmRegExp    = regexp.MustCompile("[^a-zA-Z0-9-_.]")
	imageRegExp = regexp.MustCompile("^[a-z0-9]+(([._-][a-z0-9]+)+)?$")
)

// getVMNameForDocker replicates the naming logic from Fabric's dockercontroller.GetVMNameForDocker.
// Preserved verbatim to maintain identical container name generation across Fabric versions.
func getVMNameForDocker(networkID, peerID, ccid string) (string, error) {
	// preFormatImageName
	name := ccid
	if networkID != "" && peerID != "" {
		name = fmt.Sprintf("%s-%s-%s", networkID, peerID, name)
	} else if networkID != "" {
		name = fmt.Sprintf("%s-%s", networkID, name)
	} else if peerID != "" {
		name = fmt.Sprintf("%s-%s", peerID, name)
	}
	// pre-2.0 used "-" as separator in ccid, replace ":" with "-"
	name = strings.ReplaceAll(name, ":", "-")
	sum := sha256.Sum256([]byte(name))
	hash := hex.EncodeToString(sum[:])
	saniName := vmRegExp.ReplaceAllString(name, "-")
	imageName := strings.ToLower(fmt.Sprintf("%s-%s", saniName, hash))
	if !imageRegExp.MatchString(imageName) {
		return "", fmt.Errorf("error constructing Docker VM Name: '%s' breaks Docker's repository naming rules", imageName)
	}
	return imageName, nil
}

func main() {
	netId := flag.String("net-id", "dev", "peer->networkId as specified in core.yaml")
	peerId := flag.String("peer-id", "jdoe", "peer->Id as specified in core.yaml")
	ccName := flag.String("cc-name", "ecc", "name of CC")
	ccVersion := flag.String("cc-version", "0", "version of CC")

	flag.Parse()

	// chaincode id consists of name and version, see https://github.com/hyperledger/fabric/blob/c491d69962966db1f0231496ae6cab457d8a247d/core/scc/scc.go#L24
	ccid := *ccName + ":" + *ccVersion
	name, _ := getVMNameForDocker(*netId, *peerId, ccid)
	fmt.Println(name)
}
