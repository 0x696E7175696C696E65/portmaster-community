//go:build !windows

package updates

import "testing"

func requireUpdateTestPrivileges(t *testing.T) {
	t.Helper()
}
