//go:build windows

package updates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/safing/portmaster/base/utils"
)

func TestCopyAbortsBeforeInstallOnACLFailure(t *testing.T) {
	root := t.TempDir()
	src, dst := filepath.Join(root, "source"), filepath.Join(root, "active")
	if err := os.WriteFile(src, []byte("verified content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := copyAndCheckSHA256Sum(src, dst, "", utils.FSPermission(255)); err == nil {
		t.Fatal("copy claimed success despite failed permission policy")
	}
	for _, name := range []string{dst, dst + ".copy"} {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Fatalf("unsafe output %s survived failed ACL: %v", name, err)
		}
	}
}

func TestMoveAbortsBeforeInstallOnACLFailure(t *testing.T) {
	root := t.TempDir()
	src, dst := filepath.Join(root, "staged"), filepath.Join(root, "active")
	if err := os.WriteFile(src, []byte("verified content"), 0o600); err != nil {
		t.Fatal(err)
	}
	u := &Updater{}
	if err := u.moveFile(src, dst, "", utils.FSPermission(255)); err == nil {
		t.Fatal("move claimed success despite failed permission policy")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatal("unsafe artifact became active despite failed ACL")
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatal("failed ACL lost the original staged artifact")
	}
}
