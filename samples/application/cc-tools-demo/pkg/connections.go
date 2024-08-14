/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkg

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Connections struct {
	Peers map[string]struct {
		Url string
	}

	Orderers map[string]struct {
		Url string
	}
}

func NewConnections(path string) (*Connections, error) {
	connections := &Connections{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&connections); err != nil {
		return nil, err
	}

	return connections, nil
}
