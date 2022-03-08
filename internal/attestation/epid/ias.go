/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package epid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const DefaultIASUrl = "https://api.trustedservices.intel.com/sgx/dev/attestation/v4/report"

type IntelAttestationService interface {
	RequestAttestationReport(quoteBase64 string) (reportJson string, err error)
}

type IASRequest struct {
	Quote    string `json:"isvEnclaveQuote"`
	Manifest string `json:"pseManifest,omitempty"`
	Nonce    string `json:"nonce,omitempty"`
}

type IASResponseBody struct {
	Id                    string   `json:"id"`
	Timestamp             string   `json:"timestamp"`
	Version               int      `json:"version"`
	IsvEnclaveQuoteStatus string   `json:"ISVEnclaveQuoteStatus"`
	IsvEnclaveQuoteBody   string   `json:"ISVEnclaveQuoteBody"`
	RevocationReason      string   `json:"revocationReason"`
	PseManifestStatus     string   `json:"pseManifestStatus"`
	PseManifestHash       string   `json:"pseManifestHash"`
	PlatformInfoBlob      string   `json:"platformInfoBlob"`
	Nonce                 string   `json:"nonce"`
	EpidPseudonym         string   `json:"epidPseudonym"`
	AdvisoryURL           string   `json:"advisoryURL"`
	AdvisoryIDs           []string `json:"advisoryIDs"`
}

type IASReport struct {
	Signature    string `json:"iasSignature"`
	Certificates string `json:"iasCertificates"`
	Body         string `json:"iasReport"`
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type IASClient struct {
	url        string
	apiKey     string
	httpClient HTTPClient
}

type IASClientOption func(*IASClient)

// WithUrl option allows to override the default IAS endpoint (DefaultIASUrl)
func WithUrl(url string) IASClientOption {
	return func(c *IASClient) {
		c.url = url
	}
}

// WithHttpClient option allows to use a custom http client. Mainly used for testing
func WithHttpClient(client HTTPClient) IASClientOption {
	return func(c *IASClient) {
		c.httpClient = client
	}
}

// NewIASClient returns a new IASClient instance using DefaultIASUrl as IAS endpoint
// This method requires an API Key as input in order to authenticate with the IAS.
// Optionally, IASClientOption can be provided to change the behavior of the IASClient.
func NewIASClient(apiKey string, opts ...IASClientOption) *IASClient {
	client := &IASClient{
		url:    DefaultIASUrl,
		apiKey: apiKey,
	}

	// apply options
	for _, opt := range opts {
		opt(client)
	}

	// create default http client if not provided via options
	if client.httpClient == nil {
		client.httpClient = &http.Client{}
	}

	return client
}

// RequestAttestationReport submits a quote (provided as base64 encoded string) to the Intel Attestation Service (IAS)
// in order to verify it and generate an attestation report.
// The report returned by the attestation service is packaged as a IASReport and serialized as json string.
func (i *IASClient) RequestAttestationReport(quoteBase64 string) (reportJson string, err error) {

	// build request
	request := &IASRequest{
		Quote: quoteBase64,
	}

	report, err := i.requestAttestationReport(request)
	if err != nil {
		return "", errors.Wrap(err, "")
	}

	serializedReport, err := json.Marshal(report)
	if err != nil {
		return "", errors.Wrap(err, "cannot marshal IAS report")
	}
	reportJson = string(serializedReport)

	return reportJson, nil
}

func (i *IASClient) requestAttestationReport(request *IASRequest) (report *IASReport, err error) {

	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "cannot perform http request")
	}

	// call IAS
	req, err := http.NewRequest("POST", i.url, bytes.NewReader(requestJson))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create http request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Ocp-Apim-Subscription-Key", i.apiKey)

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "cannot perform http request")
	}
	defer resp.Body.Close()

	reportRequestId := resp.Header.Get("Request-ID")

	// check response status code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed! Reason: %d %s. Request ID: %s", resp.StatusCode, resp.Status, reportRequestId)
	}

	// get header
	reportSignature := resp.Header.Get("X-IASReport-Signature")
	reportSigningCert := resp.Header.Get("X-IASReport-Signing-Certificate")

	// get the response body
	body, err := ioutil.ReadAll(resp.Body)

	report = &IASReport{
		Signature:    reportSignature,
		Certificates: reportSigningCert,
		Body:         string(body),
	}

	return report, nil
}
