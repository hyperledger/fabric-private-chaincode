/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSingleFileRejectsStrayFiles(t *testing.T) {
	dir := t.TempDir()
	if _, err := singleFile(dir); err == nil {
		t.Fatal("expected error for empty keystore")
	}

	key := filepath.Join(dir, "priv_sk")
	if err := os.WriteFile(key, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := singleFile(dir); err != nil || got != key {
		t.Fatalf("got %q, %v; want %q", got, err, key)
	}

	// hidden entries and directories, as in a mounted Kubernetes secret, are ignored
	if err := os.WriteFile(filepath.Join(dir, ".DS_Store"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "..2026_10_29"), 0o700); err != nil {
		t.Fatal(err)
	}
	if got, err := singleFile(dir); err != nil || got != key {
		t.Fatalf("got %q, %v; want %q", got, err, key)
	}

	if err := os.WriteFile(filepath.Join(dir, "old_sk"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := singleFile(dir); err == nil {
		t.Fatal("expected error for keystore with two keys")
	}
}
