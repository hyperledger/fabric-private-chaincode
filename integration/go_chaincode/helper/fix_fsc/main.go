/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {

	cmd := exec.Command("go", "mod", "download", "-json", "github.com/hyperledger-labs/fabric-smart-client")
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]string)
	err = json.Unmarshal(stdout, &m)
	if err != nil {
		log.Fatal(err)
	}

	fscPath := m["Dir"]
	builderPath := filepath.Join(fscPath, "integration", "nwo", "fabric", "fpc", "externalbuilders", "chaincode_server", "bin")

	scripts := []string{
		filepath.Join(builderPath, "build"),
		filepath.Join(builderPath, "detect"),
		filepath.Join(builderPath, "release"),
	}

	for _, s := range scripts {
		err = os.Chmod(s, 0555)
		if err != nil {
			log.Fatal(err)
		}
	}
}
