//go:build windows

package updates

import (
	"testing"

	"golang.org/x/sys/windows"
)

func requireUpdateTestPrivileges(t *testing.T) {
	t.Helper()
	// Production upgrades run as LocalSystem and intentionally protect staging
	// directories from ordinary users. These two tests install that same ACL;
	// an unelevated token cannot then write/clean their protected fixtures.
	if !windows.GetCurrentProcessToken().IsElevated() {
		t.Skip("full upgrade fixtures require an elevated Windows token; API/download security tests run without elevation")
	}
}
