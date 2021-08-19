/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/*
Notice: This file is based on fabric-sdk-go/pkg/fab/ccpackager/lifecycle/packager.go and has been modified.
*/

// Package ccpackager provides functionality to package a FPC chaincode.
//
// Example:
//  desc := &ccpackager.Descriptor{
//  	Path:    "/my_fpc_chaincode/build/_lib",
//  	Type:    ccpackager.ChaincodeType,
//  	Label:   "my-fpc-chaincode-v1",
//  	SGXMode: "SIM",
//  }
//  ccPkg, err := ccpackager.NewCCPackage(desc)
//  if err != nil {
//  	log.Fatal(err)
//  }
//
package ccpackager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/core/chaincode/persistence"
	"github.com/hyperledger/fabric/core/chaincode/platforms/util"
	"github.com/pkg/errors"
)

const (
	codePackageName          = "code.tar.gz"
	metadataPackageName      = "metadata.json"
	connectionsName          = "connection.json"
	mrenclaveFileName        = "mrenclave"
	enclaveBinaryName        = "enclave.signed.so"
	gzipCompressionLevel     = gzip.DefaultCompression
	ChaincodeType            = "fpc-c"
	CaaSType                 = "external"
	defaultConnectionTimeout = "10s"
)

// NewCCPackage creates a FPC chaincode package.
func NewCCPackage(desc *Descriptor) ([]byte, error) {
	err := desc.validate()
	if err != nil {
		return nil, err
	}

	pkgTarGzBytes, err := getTarGzBytes(desc, writePackage)
	if err != nil {
		return nil, err
	}

	return pkgTarGzBytes, nil
}

// ReadMrenclave returns mrenclave for the given FPC chaincode at ccPath
func ReadMrenclave(ccPath string) (string, error) {
	mrenclave, err := ioutil.ReadFile(filepath.Join(ccPath, mrenclaveFileName))
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(mrenclave), "\n"), nil
}

// Descriptor holds the package data. FPC supports two types of packages, ChaincodeType and CaaSType.
// For normal chaincode deployments, the package type ChaincodeType is used. It requires to define Type, Label, Path, and SGXMode.
// Alternatively, for deployments as Chaincode as a Service (CaaS), the package type CaaSType is used.
// It requires to define Type, Label, Path, and CaaSEndpoint. Optionally, CaaSTimeout and CaaSUseTLS can be set.
type Descriptor struct {
	// Type defines the FPC package type. Supported types are fpc.ChaincodeType or fpc.CaaSType.
	Type string
	// Label defines a succinct and human readable description of the package.
	Label string
	// Path defines the location of FPC enclave artifacts.
	Path string
	// SGXMode defines SGX runtime mode. Supported types are SIM and HW.
	SGXMode string
	// CaaSEndpoint defines the FPC Chaincode address if running as CaaS.
	CaaSEndpoint string
	// CaaSTimeout defines the connection timeout. Default value is 10s.
	CaaSTimeout string
	// CaaSUseTLS defines the use of TLS connection.
	CaaSUseTLS bool
}

// validate validates the package descriptor
func (p *Descriptor) validate() error {
	switch p.Type {
	case ChaincodeType:
		err := validateRegularPackageInput(p)
		if err != nil {
			return err
		}
		return nil
	case CaaSType:
		err := validateCaaSPackageInput(p)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New(fmt.Sprintf("chaincode language must be %s or %s", ChaincodeType, CaaSType))
	}
}

func validateRegularPackageInput(p *Descriptor) error {
	if p.Path == "" {
		return errors.New("chaincode path must be specified")
	}

	// TODO we should also check that the enclave artifacts exist at the given path.
	// If they do not exist, getDeploymentPayload will fail, better fail at validation earlier here.

	if p.SGXMode != sgx.SGXModeHwType && p.SGXMode != sgx.SGXModeSimType {
		return errors.Errorf("SGXMode must be set either to %s or %s, actual: %s", sgx.SGXModeHwType, sgx.SGXModeSimType, p.SGXMode)
	}

	if err := persistence.ValidateLabel(p.Label); err != nil {
		return err
	}
	return nil
}

func validateCaaSPackageInput(p *Descriptor) error {

	err := utils.ValidateEndpoint(p.CaaSEndpoint)
	if err != nil {
		return errors.Wrap(err, "CaaSEndpoint is invalid")
	}

	if err := persistence.ValidateLabel(p.Label); err != nil {
		return err
	}
	return nil
}

type writer func(tw *tar.Writer, name string, payload []byte) error

func getTarGzBytes(desc *Descriptor, writeBytesToPackage writer) ([]byte, error) {
	payload := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(payload)
	tw := tar.NewWriter(gw)

	// create metadata.json
	metadataBytes, err := metadataToJSON(desc.Path, desc.Type, desc.Label, desc.SGXMode)
	if err != nil {
		return nil, err
	}

	// write metadata.json to package
	err = writeBytesToPackage(tw, metadataPackageName, metadataBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error writing package metadata to tar")
	}

	// create code.tar.gz
	var codeBytes []byte
	switch desc.Type {
	case ChaincodeType:
		codeBytes, err = getDeploymentPayload(desc.Path)
		if err != nil {
			return nil, errors.Wrap(err, "error getting chaincode bytes")
		}
	case CaaSType:
		codeBytes, err = getCaaSDeploymentPayload(desc, writeBytesToPackage)
		if err != nil {
			return nil, errors.Wrap(err, "error getting chaincode bytes")
		}
	default:
		return nil, errors.Errorf("cannot build code.tar.gz, unknown package type = %s", desc.Type)
	}

	// write code.tar.gz to package
	err = writeBytesToPackage(tw, codePackageName, codeBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error writing package code bytes to tar")
	}

	err = tw.Close()
	if err == nil {
		err = gw.Close()
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create tar for chaincode")
	}

	return payload.Bytes(), nil
}

func writePackage(tw *tar.Writer, name string, payload []byte) error {
	err := tw.WriteHeader(
		&tar.Header{
			Name: name,
			Size: int64(len(payload)),
			Mode: 0100644,
		},
	)
	if err != nil {
		return err
	}

	_, err = tw.Write(payload)
	return err
}

func metadataToJSON(path, ccType, label, sgxMode string) ([]byte, error) {
	type packageMetadata struct {
		Path    string `json:"path,omitempty"`
		Type    string `json:"type"`
		Label   string `json:"label"`
		SGXMode string `json:"sgx_mode,omitempty"`
	}

	metadata := &packageMetadata{
		Path:    path,
		Type:    ccType,
		Label:   label,
		SGXMode: sgxMode,
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chaincode package metadata into JSON")
	}

	return metadataBytes, nil
}

func connectionToJSON(address, dialTimeout string, tlsRequired bool) ([]byte, error) {
	type connection struct {
		Address     string `json:"address"`
		DialTimeout string `json:"dial_timeout"`
		TLSRequired bool   `json:"tls_required"`
	}

	connections := &connection{
		Address:     address,
		DialTimeout: dialTimeout,
		TLSRequired: tlsRequired,
	}

	connectionsBytes, err := json.Marshal(connections)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal connections into JSON")
	}

	return connectionsBytes, nil

}
func getDeploymentPayload(ccPath string) ([]byte, error) {
	type file struct {
		Path string
		Name string
	}

	// FPC code package (code.tar.gz) only contains mrenclave file and enclave.signed.so
	files := []file{
		{ccPath, mrenclaveFileName},
		{ccPath, enclaveBinaryName},
	}

	payload := bytes.NewBuffer(nil)
	gw, err := gzip.NewWriterLevel(payload, gzipCompressionLevel)
	if err != nil {
		return nil, err
	}
	tw := tar.NewWriter(gw)

	for _, file := range files {
		err = util.WriteFileToPackage(filepath.Join(file.Path, file.Name), file.Name, tw)
		if err != nil {
			return nil, errors.Wrapf(err, "error writing %s to tar", file.Name)
		}
	}

	err = tw.Close()
	if err == nil {
		err = gw.Close()
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create code.tar.gz for chaincode")
	}

	return payload.Bytes(), nil
}

func getCaaSDeploymentPayload(desc *Descriptor, writeBytesToPackage writer) ([]byte, error) {

	// set default timeout
	if desc.CaaSTimeout == "" {
		desc.CaaSTimeout = defaultConnectionTimeout
	}

	connectionBytes, err := connectionToJSON(desc.CaaSEndpoint, desc.CaaSTimeout, desc.CaaSUseTLS)
	if err != nil {
		return nil, err
	}

	payload := bytes.NewBuffer(nil)
	gw, err := gzip.NewWriterLevel(payload, gzipCompressionLevel)
	if err != nil {
		return nil, err
	}
	tw := tar.NewWriter(gw)

	err = writeBytesToPackage(tw, connectionsName, connectionBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "error writing %s to tar", connectionsName)
	}

	err = tw.Close()
	if err == nil {
		err = gw.Close()
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create code.tar.gz for chaincode")
	}

	return payload.Bytes(), nil
}
