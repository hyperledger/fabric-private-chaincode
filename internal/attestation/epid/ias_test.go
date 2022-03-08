/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package epid

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/fakes"
	"github.com/stretchr/testify/assert"
)

//go:generate counterfeiter -o fakes/httpclient.go -fake-name HTTPClient . httpClient
//lint:ignore U1000 This is just used to generate fake
type httpClient interface {
	HTTPClient
}

func TestIAS(t *testing.T) {
	dummyQuote := base64.StdEncoding.EncodeToString([]byte("dummyQuote"))
	dummyApiKey := "some_key"
	dummyIASUrl := "https://api.fakeias.com/attestation/report"

	expectedSignature := "signature"
	expectedCertificates := "certs"
	expectedBody := "some body"

	expectedReport := &IASReport{
		Signature:    expectedSignature,
		Certificates: expectedCertificates,
		Body:         expectedBody,
	}

	header := http.Header{}
	header.Add("X-IASReport-Signature", expectedSignature)
	header.Add("X-IASReport-Signing-Certificate", expectedCertificates)

	fakeHttpClient := &fakes.HTTPClient{}
	fakeHttpClient.DoReturns(&http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader(expectedBody)),
	}, nil)

	iasClient := NewIASClient(dummyApiKey, WithHttpClient(fakeHttpClient), WithUrl(dummyIASUrl))
	assert.Equal(t, dummyApiKey, iasClient.apiKey)
	assert.Equal(t, dummyIASUrl, iasClient.url)
	assert.NotNil(t, iasClient.httpClient)

	reportJson, err := iasClient.RequestAttestationReport(dummyQuote)
	assert.NoError(t, err)
	assert.NotEmpty(t, reportJson)

	httpReq := fakeHttpClient.DoArgsForCall(0)
	assert.Equal(t, "POST", httpReq.Method)
	assert.Equal(t, "application/json", httpReq.Header.Get("Content-Type"))
	assert.Equal(t, dummyApiKey, httpReq.Header.Get("Ocp-Apim-Subscription-Key"))
	assert.Equal(t, dummyIASUrl, httpReq.URL.String())

	report := &IASReport{}
	err = json.Unmarshal([]byte(reportJson), report)
	assert.NoError(t, err)

	assert.EqualValues(t, expectedReport, report)
}
