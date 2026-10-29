/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ccpackager

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteFileToPackageIsDeterministic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}

	pack := func(mtime time.Time) []byte {
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		if err := writeFileToPackage(path, "src/file", tw); err != nil {
			t.Fatal(err)
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}

	if !bytes.Equal(pack(time.Unix(1000, 0)), pack(time.Unix(2000, 0))) {
		t.Fatal("package content depends on file modification time")
	}
}
