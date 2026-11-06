/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const sep = "."

func Read(file string) []byte {
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	if data == nil {
		panic(fmt.Errorf("file is empty"))
	}
	return data
}

func IsFPCCompositeKey(comp string) bool {
	return len(comp) >= 2 && strings.HasPrefix(comp, sep) && strings.HasSuffix(comp, sep)
}

func TransformToFPCKey(comp string) string {
	return strings.Replace(comp, "\x00", sep, -1)
}

func SplitFPCCompositeKey(comp_str string) []string {
	// check it has sep in front and end
	if !IsFPCCompositeKey(comp_str) {
		panic("comp_key has wrong format")
	}
	comp := strings.Split(comp_str, sep)
	return comp[1 : len(comp)-1]
}

// compositeKeyCreator is the part of shim.ChaincodeStubInterface that ToFabricKey needs.
type compositeKeyCreator interface {
	CreateCompositeKey(objectType string, attributes []string) (string, error)
}

// ToFabricKey returns the ledger key for an FPC key. An FPC composite key is
// turned back into a Fabric composite key; any other key is returned unchanged.
func ToFabricKey(stub compositeKeyCreator, key string) (string, error) {
	if !IsFPCCompositeKey(key) {
		return key, nil
	}
	comp := SplitFPCCompositeKey(key)
	return stub.CreateCompositeKey(comp[0], comp[1:])
}

func ValidateEndpoint(endpoint string) error {
	colon := strings.LastIndexByte(endpoint, ':')
	if colon == -1 {
		return fmt.Errorf("invalid format")
	}

	_, err := strconv.Atoi(endpoint[colon+1:])
	if err != nil {
		return errors.Wrap(err, "invalid port")
	}

	return nil
}
