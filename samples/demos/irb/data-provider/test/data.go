/*
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
)

var dataDir string = "data"

func CreateData(dataFileName string) error {
	err := os.MkdirAll(filepath.Join(dataDir), 0755)
	if err != nil {
		return err
	}

	n1 := rand.Intn(20) + 30 // random n between 30 - 50
	n2 := rand.Intn(10)      // random n between 0 - 10

	data := fmt.Sprintf("%d.%d, 0, 0, 1, 1, 0", n1, n2)

	err = ioutil.WriteFile(filepath.Join(dataDir, dataFileName), []byte(data), 0755)
	if err != nil {
		return err
	}

	return nil
}

func LoadData(dataFileName string) (data []byte, e error) {
	data, err := ioutil.ReadFile(filepath.Join(dataDir, dataFileName))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func LoadOrCreateData(dataFileName string) (data []byte, e error) {
	d, err := LoadData(dataFileName)
	if err == nil {
		return d, nil
	}

	err = CreateData(dataFileName)
	if err != nil {
		return nil, err
	}

	return LoadData(dataFileName)
}
