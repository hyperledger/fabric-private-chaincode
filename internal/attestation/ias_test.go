/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/fakes"
	"github.com/stretchr/testify/assert"
)

//go:generate counterfeiter -o fakes/httpclient.go -fake-name HTTPClient . httpClient
type httpClient interface {
	HTTPClient
}

func TestIAS(t *testing.T) {

	fakeHttpClient := &fakes.HTTPClient{}

	expectedApiKey := "some_key"
	iasClient := NewIASClient(expectedApiKey, WithHttpClient(fakeHttpClient))
	assert.Equal(t, expectedApiKey, iasClient.apiKey)
	assert.Equal(t, DefaultIASUrl, iasClient.url)
	assert.NotNil(t, iasClient.httpClient)

	//iasClient.RequestAttestationReport()

}
