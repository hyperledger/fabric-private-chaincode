/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package attestation

import (
	"encoding/base64"
	"testing"
)

const enclavePK = `MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEE9lPD9QkW9oxWlFvwABrmseYAVvoBvvmTt3jzV0sdASR2KDDQPvz8EcyqfomEOTwSz7E+mISktMxYqofRr+4Yw==`
const enclavePkHash = `qpEqqBaEkNz9bTO77QK8+CLbvaEN1NATs7ajRTzq70k=`
const quote = `AgAAAG4NAAAEAAQAAAAAACVC+Q1jMSwdovbiGHbw44nMDb+CvAvF0FJF/38NWjOqAgIC/wEBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABwAAAAAAAAAHAAAAAAAAAJiu1hyR8lijfGjtSUMpdpVkfse75gCMwRGwoSZQ6+uRAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACD1xnnferKFHD2uvYqTXdDA8iZ22kCD5xw7h38CMfOngAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD3TvjLWa36sT/kCIRYXhtYoRQ61x2u48Q16bzoq8w6egAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAqAIAAFXdt6ObnofTxKVhK9Eafot/LsUGgr4546W34JUey7aqo9b6mpeP3X6W/DMIc1JbIXpFd5+mHWP+R7swgSEYNg+RUVbjkZ38nOJHzIl0E7Dgxs8X8iilH+hxpcPiYQphpIcBS5NUCmDn6Wsz/I+Dbpbt3e2G74WFPLHqDn+JHva5vtaYHd7cAfmPIhZZXXMCQJ8um5Jcer4L16VOugt8LEE0i3FqLb0khMYUHmEsqWuh1Fss5bNUuRDqotz6XTBq0uQ+nCfzv9ZsT2CDihsuQTzgU0BiZZuf06Aw9NQdywg+vTZoqyWw0Ca/jsAt+OpbQeQzDoH3HAvnaRvRByozHqKQ1Z83vVny2DQPVwWm6hxIEUCDVE2A/fkbo+UjR12fD8XWUw3xXfd6Dob9N2gBAAB1NRKH8uhAp94KvF/EF76xtBOYnlpAkbv4pYsmJfWkt0CtKtt/lvMQqkmZwSi8LQ93XBiAdVEKt255ycfFxcmAHPFPrjwHMb0/5wKNXa9vyBlgJ63tU/8U1JxujZ6QdS05xiQbKb+l2y6Nm++iw1Ba7BBJgQR+xDBud/VMjjLI3/nMlA9JTpVw9sSTsWdqHzA4bJm2P7fxkxL4wUYe6w+1uWGnT8XFwuJOfw1bUKZWlGZCOe8iLiPmDOmKUegpiLy0wY73gk+5bJhq1L8b4EXJMoSVoS4JgzYajh8oEBaUheiR4ze8sD9KuF0y+dfQklcMKdONyXMcI8QcZfj19iQy2FvXY8Ca0AoBkQMk4bn49e19ePChDUhrk7ynGGy5d9Wo8g3aNZLNWol5LuwCduTYv83xbHeKDkEsvk23m5NiXlVnDo6Pwu+32w57sX4K4CcojQZvJRYfFUuRCoN05TY0oJ0qvvZ1pAEAQAuBfOucbX6QZZ4qPcMR`

// TODO: below might have to be fixed for changes related to new IAS auth method (issue #47, PR #49).
// However, i have no idea where NewIASCredentialProviderFromConfig comes from and
// whether this function is anywhere called/supposed to work and it works without any changes, so probably dead code ...?
func TestRequestAttestationReport(t *testing.T) {

	ias := NewIAS()
	credis, _ := NewIASCredentialProviderFromConfig()
	verifier := VerifierImpl{}

	quoteAsBytes, err := base64.StdEncoding.DecodeString(quote)
	if err != nil {
		jsonResp := "{\"Error\":\" Can not parse quoteBase64 string: " + err.Error() + " \"}"
		t.Errorf(jsonResp)
	}

	// get ercc client cert for IAS
	cert, err := credis.GetIASClientCert()
	if err != nil {
		jsonResp := "{\"Error\":\" Can not retrieve IAS client cert from ledger: " + err.Error() + " \"}"
		t.Errorf(jsonResp)
	}

	// send quote to intel for verification
	attestationReport, err := ias.RequestAttestationReport(cert, quoteAsBytes)
	if err != nil {
		jsonResp := "{\"Error\":\" Error while retrieving attestation report: " + err.Error() + "\"}"
		t.Errorf(jsonResp)
	}

	verificationPK, err := ias.GetIntelVerificationKey()
	if err != nil {
		t.Errorf("Can not parse verifiaction key: %s", err)
	}

	// verify attestation report
	isValid, err := verifier.VerifyAttestionReport(verificationPK, attestationReport)
	if err != nil {
		jsonResp := "{\"Error\":\" Error while attestation report verification: " + err.Error() + "\"}"
		t.Errorf(jsonResp)
	}
	if !isValid {
		jsonResp := "{\"Error\":\" Attestation report is not valid \"}"
		t.Errorf(jsonResp)
	}

}
