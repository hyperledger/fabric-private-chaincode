/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	quote  = "dummyQuote"
	apiKey = "dummyAPiKey"
)

var _ = Describe("Ias", func() {

	When("invoke GetIntelVerificationKey", func() {
		It("should return Intel pub key", func() {
			ias := attestation.NewIAS()
			pem, err := ias.GetIntelVerificationKey()
			Expect(err).NotTo(HaveOccurred())
			Expect(pem).NotTo(BeNil())
		})
	})

	Context("invoke RequestAttestationReport", func() {

		When("IAS returns an error", func() {
			It("should return an error", func() {
				ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Wrong API ", http.StatusUnauthorized)

				}))
				defer ts.Close()

				ias := attestation.NewIASWithMock(ts.URL, ts.Client())
				_, err := ias.RequestAttestationReport(apiKey, []byte(quote))
				Expect(err).To(MatchError("IAS returned error: Code 401 Unauthorized"))
			})
		})

		When("response does not contain return IASReportBody", func() {
			It("should return an error", func() {
				ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("NOT a IASReportBody"))
				}))
				defer ts.Close()

				ias := attestation.NewIASWithMock(ts.URL, ts.Client())
				_, err := ias.RequestAttestationReport(apiKey, []byte(quote))
				Expect(err).To(HaveOccurred())
				Expect(strings.HasPrefix(err.Error(), "cannot unmarshal report body:")).To(BeTrue())
			})
		})

		When("IASReportBody does not contain submitted quote", func() {
			It("should return an error", func() {
				emptyReport := attestation.IASReportBody{}
				emptyReportBytes, err := json.Marshal(emptyReport)
				Expect(err).NotTo(HaveOccurred())

				ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write(emptyReportBytes)
				}))
				defer ts.Close()

				ias := attestation.NewIASWithMock(ts.URL, ts.Client())
				_, err = ias.RequestAttestationReport(apiKey, []byte(quote))
				Expect(err).To(HaveOccurred())
				Expect(strings.HasPrefix(err.Error(), "report does not contain submitted quote")).To(BeTrue())
			})
		})

		When("response is OK", func() {
			It("should return return a non-empty IASAttestationReport", func() {
				quoteBase64 := base64.StdEncoding.EncodeToString([]byte(quote))
				report := attestation.IASReportBody{IsvEnclaveQuoteBody: quoteBase64}
				reportBytes, err := json.Marshal(report)
				Expect(err).NotTo(HaveOccurred())

				ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("X-IASReport-Signature", "sig")
					w.Header().Add("X-IASReport-Signing-Certificate", "cert")
					_, _ = w.Write(reportBytes)
				}))
				defer ts.Close()

				ias := attestation.NewIASWithMock(ts.URL, ts.Client())
				res, err := ias.RequestAttestationReport(apiKey, []byte(quote))
				Expect(err).NotTo(HaveOccurred())
				Expect(res.IASReportBody).To(MatchJSON(reportBytes))
				Expect(res.IASReportSignature).To(Equal("sig"))
				Expect(res.IASReportSigningCertificate).To(Equal("cert"))
			})
		})

	})

})
