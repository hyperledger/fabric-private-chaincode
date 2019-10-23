/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var MockData map[string]interface{}

func init() {
	MockData, _ = loadMockData()
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
