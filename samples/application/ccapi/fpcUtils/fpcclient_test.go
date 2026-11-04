/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpcUtils

import "testing"

func TestNewClientReturnsError(t *testing.T) {
	_, err := NewClient(&Config{CorePeerTLSEnabled: true, CorePeerTLSRootCertFile: "does-not-exist"})
	if err == nil {
		t.Fatal("expected an error for a missing TLS root certificate")
	}
}
