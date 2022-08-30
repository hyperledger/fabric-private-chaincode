package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func DetectSgxDevicePath() ([]string, error) {
	possiblePaths := []string{"/dev/isgx", "/dev/sgx/enclave"}
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err != nil {
			continue
		} else {
			// first found path returns
			return []string{p}, nil
		}
	}

	return nil, fmt.Errorf("no sgx device path found")
}

func ReadMrenclaveFromFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read mrenclave from %s", path)
	}

	mrenclave := strings.TrimSpace(string(data))
	if len(mrenclave) == 0 {
		return "", fmt.Errorf("mrenclave file empty")
	}

	return mrenclave, nil
}
