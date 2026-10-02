//go:build !linux && !windows

package proxy

import (
	"fmt"
	"net"
	"syscall"
)

// An unsupported platform must refuse a selected interface rather than
// silently sending privacy-sensitive traffic over its default route.
func applyBindToDevice(d *net.Dialer, iface string) {
	if iface == "" {
		return
	}
	d.Control = func(_, _ string, _ syscall.RawConn) error {
		return fmt.Errorf("interface binding is unsupported on this platform: %q", iface)
	}
}
