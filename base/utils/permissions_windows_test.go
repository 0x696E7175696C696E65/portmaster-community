//go:build windows

package utils

import (
	"path/filepath"
	"testing"
)

// Security-sensitive callers must be told when an ACL could not be installed.
func TestSetFilePermissionReportsACLFailure(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	for _, permission := range []FSPermission{AdminOnlyPermission, PublicReadPermission, PublicWritePermission} {
		if err := SetFilePermission(missing, permission); err == nil {
			t.Fatalf("permission %d silently ignored an ACL failure", permission)
		}
	}
}

func TestSetFilePermissionRejectsUnknownPermission(t *testing.T) {
	if err := SetFilePermission(t.TempDir(), FSPermission(255)); err == nil {
		t.Fatal("unknown permission must not report success")
	}
}

func TestEnsureDirectoryPropagatesPermissionError(t *testing.T) {
	// An invalid policy lets us exercise the failure path without requiring
	// elevation or changing access to the test runner's temporary directory.
	for _, path := range []string{t.TempDir(), filepath.Join(t.TempDir(), "new")} {
		if err := EnsureDirectory(path, FSPermission(255)); err == nil {
			t.Fatal("directory creation silently ignored a permission failure")
		}
	}
}
