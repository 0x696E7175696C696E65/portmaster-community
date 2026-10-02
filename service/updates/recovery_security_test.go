package updates

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFailedRollbackReportsEveryFailedRestore(t *testing.T) {
	root := t.TempDir()
	purge, active := filepath.Join(root, "purge"), filepath.Join(root, "active")
	if err := os.Mkdir(purge, 0o700); err != nil {
		t.Fatal(err)
	}
	// An active installation path occupied by a file cannot receive resources.
	if err := os.WriteFile(active, []byte("blocking destination"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"first.dat", "second.dat"} {
		if err := os.WriteFile(filepath.Join(purge, name), []byte("backup"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	u := &Updater{cfg: Config{Name: "rollback test", PurgeDirectory: purge, Directory: active}}
	err := u.recoverFromFailedUpgrade()
	if err == nil || !strings.Contains(err.Error(), "first.dat") || !strings.Contains(err.Error(), "second.dat") {
		t.Fatalf("failed restores incorrectly reported success or lost errors: %v", err)
	}
	for _, name := range []string{"first.dat", "second.dat"} {
		if _, err := os.Stat(filepath.Join(purge, name)); err != nil {
			t.Fatalf("failed rollback lost backup %s: %v", name, err)
		}
	}
}

func TestFailedPurgePreparationLeavesActiveFilesUntouched(t *testing.T) {
	root := t.TempDir()
	active := filepath.Join(root, "active")
	if err := os.Mkdir(active, 0o700); err != nil {
		t.Fatal(err)
	}
	current := filepath.Join(active, "current.dat")
	if err := os.WriteFile(current, []byte("current installation"), 0o600); err != nil {
		t.Fatal(err)
	}
	blockedParent := filepath.Join(root, "blocked")
	if err := os.WriteFile(blockedParent, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	u := &Updater{cfg: Config{Name: "preparation test", Directory: active, PurgeDirectory: filepath.Join(blockedParent, "purge")}}
	err := u.upgrade(&Downloader{index: &Index{Version: "1.0.1"}}, true)
	if !errors.Is(err, errUpgradePreparation) || strings.Contains(err.Error(), "recovery was successful") {
		t.Fatalf("preparation error was not reported accurately: %v", err)
	}
	if data, err := os.ReadFile(current); err != nil || string(data) != "current installation" {
		t.Fatal("failed preparation modified the active installation")
	}
}
