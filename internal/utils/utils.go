/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
)

const MrEnclaveStateKey = "MRENCLAVE"

// Response contains the response data and signature produced by the enclave
// TODO replace this with a proto? TBD
type Response struct {
	ResponseData []byte `json:"ResponseData"`
	Signature    []byte `json:"Signature"`
	PublicKey    []byte `json:"PublicKey"`
	//	TODO add read/write set
}

func UnmarshalResponse(respBytes []byte) (*Response, error) {
	response := &Response{}
	err := json.Unmarshal(respBytes, response)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling FPC response err: %s", err)
	}
	return response, nil
}

// TODO replace this with a proto? TBD
type ChaincodeParams struct {
	Function            string   `json:"Function"`
	Args                []string `json:"args"`
	ResultEncryptionKey []byte   `json:"ResultEncryptionKey"`
}

// TODO replace this with a proto? TBD
type AttestationParams struct {
	Params []string `json:"params"`
}

const sep = "."

func Read(file string) []byte {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	if data == nil {
		panic(fmt.Errorf("File is empty"))
	}
	return data
}

func IsFPCCompositeKey(comp string) bool {
	return strings.HasPrefix(comp, sep) && strings.HasSuffix(comp, sep)
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

// returns enclave_id as hex-encoded string of SHA256 hash over enclave_vk.
func GetEnclaveId(attestedData *protos.AttestedData) string {
	h := sha256.Sum256(attestedData.EnclaveVk)
	return hex.EncodeToString(h[:])
}
