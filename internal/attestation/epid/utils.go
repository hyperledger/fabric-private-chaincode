/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package epid

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// loadApiKey tries to load the IAS API Key from environment variable.
// If env var not set, use loadApiKeyFromCredentialsEnvPath and then loadApiKeyFromFPCConfig as fallback
func loadApiKey() (string, error) {
	// try to load from env variable
	apiKey := os.Getenv("IAS_API_KEY")
	if len(apiKey) != 0 {
		return apiKey, nil
	}

	// fallback read from $SGX_CREDENTIALS_PATH
	apiKey, err := loadApiKeyFromCredentialsEnvPath()
	if err == nil {
		return apiKey, nil
	}

	// fallback read from $FPC_PATH
	return loadApiKeyFromFPCConfig()
}

func loadApiKeyFromCredentialsEnvPath() (string, error) {
	path := os.Getenv("SGX_CREDENTIALS_PATH")
	if len(path) == 0 {
		return "", fmt.Errorf("$SGX_CREDENTIALS_PATH not set")
	}

	return loadApiKeyFromPath(path)
}

// loadApiKeyFromFPCConfig tries to load IAS API Key from $FPC_PATH/config/ias/api_key.txt
func loadApiKeyFromFPCConfig() (string, error) {
	fpcPath := os.Getenv("FPC_PATH")
	if len(fpcPath) == 0 {
		return "", fmt.Errorf("$FPC_PATH not set")
	}

	path := filepath.Join(fpcPath, "config", "ias")
	return loadApiKeyFromPath(path)
}

// loadApiKeyFromPath loads IAS API key from path/api_key.txt
func loadApiKeyFromPath(path string) (string, error) {
	path = filepath.Join(path, "api_key.txt")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "could not read %s", path)
	}

	if len(data) == 0 {
		return "", errors.Errorf("empty file %s", path)
	}

	return strings.TrimSuffix(string(data), "\n"), nil
}
