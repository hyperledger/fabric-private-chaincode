/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2019 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hyperledger/fabric/common/flogging"
)

// intel verification key
const IntelPubPEM = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqXot4OZuphR8nudFrAFi
aGxxkgma/Es/BA+tbeCTUR106AL1ENcWA4FX3K+E9BBL0/7X5rj5nIgX/R/1ubhk
KWw9gfqPG3KeAtIdcv/uTO1yXv50vqaPvE1CRChvzdS/ZEBqQ5oVvLTPZ3VEicQj
lytKgN9cLnxbwtuvLUK7eyRPfJW/ksddOzP8VBBniolYnRCD2jrMRZ8nBM2ZWYwn
XnwYeOAHV+W9tOhAImwRwKF/95yAsVwd21ryHMJBcGH70qLagZ7Ttyt++qO/6+KA
XJuKwZqjRlEtSEz8gZQeFfVYgcwSfo96oSMAzVr7V0L6HSDLRnpb6xxmbPdqNol4
tQIDAQAB
-----END PUBLIC KEY-----`

const iasURL = "https://api.trustedservices.intel.com/sgx/dev/attestation/v3/report"

var logger = flogging.MustGetLogger("ercc.ias")

// IASReportBody received from IAS (Intel attestation service)
type IASReportBody struct {
	ID                    string `json:"id"`
	IsvEnclaveQuoteStatus string `json:"isvEnclaveQuoteStatus"`
	IsvEnclaveQuoteBody   string `json:"isvEnclaveQuoteBody"`
	PlatformInfoBlob      string `json:"platformInfoBlob,omitempty"`
	RevocationReason      string `json:"revocationReason,omitempty"`
	PseManifestStatus     string `json:"pseManifestStatus,omitempty"`
	PseManifestHash       string `json:"pseManifestHash,omitempty"`
	Nonce                 string `json:"nonce,omitempty"`
	EpidPseudonym         string `json:"epidPseudonym,omitempty"`
	Timestamp             string `json:"timestamp"`
}

// IASAttestationReport received from IAS (Intel attestation service)
// TODO renamte to AttestationReport
type IASAttestationReport struct {
	EnclavePk                   []byte `json:"EnclavePk"`
	IASReportSignature          string `json:"IASReport-Signature"`
	IASReportSigningCertificate string `json:"IASReport-Signing-Certificate"`
	IASReportBody               []byte `json:"IASResponseBody"`
}

// IntelAttestationService sent to IAS (Intel attestation service)
type IntelAttestationService interface {
	RequestAttestationReport(apiKey string, quoteAsBytes []byte) (IASAttestationReport, error)
	GetIntelVerificationKey() (interface{}, error)
}

// NewIAS is a great help to build an IntelAttestationService object
func NewIAS() IntelAttestationService {
	return &intelAttestationServiceImpl{
		url:    iasURL,
		client: setupClient(),
	}
}

func NewIASWithMock(mockURL string, mockClient *http.Client) IntelAttestationService {
	return &intelAttestationServiceImpl{
		url:    mockURL,
		client: mockClient,
	}
}

type intelAttestationServiceImpl struct {
	url    string
	client *http.Client
}

func setupClient() *http.Client {
	// Setup HTTPS client
	tlsConfig := &tls.Config{
		// RootCAs:            caCertPool,
		InsecureSkipVerify: true, // TODO: fix this. with api-keys we really should verify IAS ...
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}
	return &http.Client{Transport: transport}
}

// RequestAttestationReport sends a quote to Intel for verification and in return receives an IASAttestationReport
// Calling Intel qualifies ercc as a system chaincode since in the future chaincodes might be restricted and can not make call outside their docker container
func (ias *intelAttestationServiceImpl) RequestAttestationReport(apiKey string, quoteAsBytes []byte) (IASAttestationReport, error) {

	// transform quote bytes to base64 and build request body
	quoteAsBase64 := base64.StdEncoding.EncodeToString(quoteAsBytes)
	requestBody := &IASRequestBody{Quote: quoteAsBase64}
	requestBytes, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", ias.url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return IASAttestationReport{}, fmt.Errorf("IAS connection error: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Ocp-Apim-Subscription-Key", apiKey)
	logger.Debugf("Sending IAS request %s", req)

	// submit quote for verification
	resp, err := ias.client.Do(req)
	if err != nil {
		return IASAttestationReport{}, fmt.Errorf("IAS connection error: %s", err)
	}
	defer resp.Body.Close()

	logger.Debugf("Received IAS response %s", resp)

	// check response
	if resp.StatusCode != 200 {
		return IASAttestationReport{}, fmt.Errorf("IAS returned error: Code %s", resp.Status)
	}

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return IASAttestationReport{}, fmt.Errorf("cannot read response body: %s", err)
	}

	reportBody := IASReportBody{}
	err = json.Unmarshal(bodyData, &reportBody)
	if err != nil {
		return IASAttestationReport{}, fmt.Errorf("cannot unmarshal report body: %s", err)
	}

	// check response contains submitted quote.
	// That way, we can be sure (once signature check later works out) that it is
	// a response to our request, even if there is some TLS issue (note we are using the
	// general TLS trust-store, so a multitude of CAs would be trusted and could compromise
	// security)
	// Note: ias does not return the complete quotebody but skips EPID signature and signature length,
	//   we have to do prefix rather than equality match (better would even be to decode base64
	//   and check that first 432 bytes match ...)
	if len(reportBody.IsvEnclaveQuoteBody) == 0 || !strings.HasPrefix(quoteAsBase64, reportBody.IsvEnclaveQuoteBody) {
		return IASAttestationReport{}, fmt.Errorf("report does not contain submitted quote (%s not properly prefixed by %s)", quoteAsBase64, reportBody.IsvEnclaveQuoteBody)
	}

	report := IASAttestationReport{
		IASReportSignature:          resp.Header.Get("X-IASReport-Signature"),
		IASReportSigningCertificate: resp.Header.Get("X-IASReport-Signing-Certificate"),
		IASReportBody:               bodyData,
	}

	return report, nil
}

func (ias *intelAttestationServiceImpl) GetIntelVerificationKey() (interface{}, error) {
	return PublicKeyFromPem([]byte(IntelPubPEM))
}

func PublicKeyFromPem(bytes []byte) (interface{}, error) {
	block, _ := pem.Decode([]byte(bytes))
	if block == nil {
		return nil, fmt.Errorf("Failed to parse PEM block containing the public key")
	}
	pk, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Public key is invalid: %s", err)
	}
	return pk, nil
}
