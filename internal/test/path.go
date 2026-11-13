package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func projectRoot(t testing.TB) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get test working directory: %v", err)
	}
	// Locate the module root without changing the working directory of other tests.
	for {
		_, err = os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil {
			return dir
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("locate project root: %v", err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("locate project root: go.mod not found")
		}
		dir = parent
	}
}
