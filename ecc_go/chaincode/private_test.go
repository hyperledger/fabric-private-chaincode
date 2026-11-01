/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import "testing"

func TestWarnIfNoSGX(t *testing.T) {
	const want = "FPC running without SGX: no confidentiality, for development only"
	for _, tc := range []struct {
		noSGX, oeSimulation string
		warns               bool
	}{
		{noSGX: "true", oeSimulation: "", warns: true}, // plain go build (with_go)
		{noSGX: "", oeSimulation: "1", warns: true},    // ego in SGX simulation mode
		{noSGX: "", oeSimulation: "", warns: false},    // ego in SGX hardware mode
		{noSGX: "false", oeSimulation: "0", warns: false},
	} {
		var got []string
		warnIfNoSGX(tc.noSGX, tc.oeSimulation, func(msg string) { got = append(got, msg) })

		if !tc.warns {
			if len(got) != 0 {
				t.Errorf("noSGX=%q oeSimulation=%q: unexpected warning %q", tc.noSGX, tc.oeSimulation, got)
			}
			continue
		}
		if len(got) != 1 || got[0] != want {
			t.Errorf("noSGX=%q oeSimulation=%q: got %q, want [%q]", tc.noSGX, tc.oeSimulation, got, want)
		}
	}
}
