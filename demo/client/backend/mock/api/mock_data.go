/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type MappedName struct {
	User  string `json:"user"`
	MspId string `json:"mspId"`
	Org   string `json:"org"`
}

var MockData map[string]interface{}
var MockNameMap map[string]MappedName

func init() {
	var err error
	if MockData, err = loadMockData(); err != nil {
		panic(fmt.Sprintf("Cannot read mock data: error=%v", err))
	}
	if MockNameMap, err = loadMockNameMap(); err != nil {
		panic(fmt.Sprintf("Cannot read mock name mapping: error=%v", err))
	}
}

func loadMockData() (map[string]interface{}, error) {
	jsonFile, err := os.Open("api/serverapi.json")
	defer jsonFile.Close()
	if err != nil {
		return nil, err
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func loadMockNameMap() (map[string]MappedName, error) {
	jsonFile, err := os.Open("api/name-map.json")
	defer jsonFile.Close()
	if err != nil {
		return nil, err
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var result map[string]MappedName
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
