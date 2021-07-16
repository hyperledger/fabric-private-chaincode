/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package sgx provides Intel SGX specific functionality.
package sgx

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	SGXModeEnvKey         = "SGX_MODE"
	SGXModeHwType         = "HW"
	SGXModeSimType        = "SIM"
	SGXCredentialsPathKey = "SGX_CREDENTIALS_PATH"
)

// AttestationParams holds additional attestation information that is required to perform LifecycleInitEnclave.
type AttestationParams struct {
	AttestationType string `json:"attestation_type"`
	HexSpid         string `json:"hex_spid"`
	SigRL           string `json:"sig_rl"`
}

// ToBase64EncodedJSON returns the SGXAttestationParams object as serialized JSON with Base64 encoding.
func (p *AttestationParams) ToBase64EncodedJSON() ([]byte, error) {
	serializedParams, err := json.Marshal(p)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot marshall (json) attestation params '%v'", p)
	}

	return []byte(base64.StdEncoding.EncodeToString(serializedParams)), nil
}

// Validate checks that the attestation information are correct.
func (p *AttestationParams) Validate() error {
	// TODO implement me
	// check that attestation type is correct (epid, cap, simulated)
	// check that hexspid has right format? etc ...

	return nil
}

// CreateAttestationParamsFromEnvironment reads attestation information from environment variables and returns
// an SGXAttestationParams object. This methods requires `SGX_MODE` and `SGX_CREDENTIALS_PATH` to be set.
func CreateAttestationParamsFromEnvironment() (*AttestationParams, error) {
	switch sgxMode := os.Getenv(SGXModeEnvKey); sgxMode {
	case SGXModeHwType:
		sgxCredentialsPath := os.Getenv(SGXCredentialsPathKey)
		if sgxCredentialsPath == "" {
			return nil, errors.Errorf("%s environment variable undefined", SGXCredentialsPathKey)
		}
		return CreateAttestationParamsFromCredentialsPath(sgxCredentialsPath)

	case SGXModeSimType:
		return &AttestationParams{
			AttestationType: "simulated",
		}, nil

	default:
		return nil, errors.Errorf("%s environment variable ill-defined: '%s'", SGXModeEnvKey, sgxMode)
	}
}

// CreateAttestationParamsFromCredentialsPath reads attestation information from a given path and returns an
// SGXAttestationParams object.
func CreateAttestationParamsFromCredentialsPath(sgxCredentialsPath string) (*AttestationParams, error) {
	spidType, err := ReadSPIDType(sgxCredentialsPath)
	if err != nil {
		return nil, err
	}

	hexSpid, err := ReadSPID(sgxCredentialsPath)
	if err != nil {
		return nil, err
	}

	sigRL, err := ReadSigRL(sgxCredentialsPath)
	if err != nil {
		return nil, err
	}

	return &AttestationParams{
		AttestationType: spidType,
		HexSpid:         hexSpid,
		SigRL:           sigRL,
	}, nil

}

// ReadSPIDType reads the SPID type from a credentials path and returns it as string.
func ReadSPIDType(sgxCredentialsPath string) (string, error) {
	spidTypePath := filepath.Join(sgxCredentialsPath, "spid_type.txt")
	return readFile(spidTypePath)
}

// ReadSPID reads the SPID from a credentials path and returns it as string.
func ReadSPID(sgxCredentialsPath string) (string, error) {
	hexSpidPath := filepath.Join(sgxCredentialsPath, "spid.txt")
	return readFile(hexSpidPath)
}

// ReadSigRL reads the Signature Revocation List from a credentials path and returns it as string.
func ReadSigRL(sgxCredentialsPath string) (string, error) {
	// TODO implement me
	return "", nil
}

func readFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "could not read %s", path)
	}

	if len(content) == 0 {
		return "", errors.Errorf("empty file %s", path)
	}

	return strings.TrimSuffix(string(content), "\n"), nil
}
