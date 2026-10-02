//go:build windows

package proxy

import (
	"encoding/binary"
	"net"
	"syscall"

	"golang.org/x/sys/windows"
)

// Restrict WireGuard egress to the selected interface in addition to binding
// its source IP. A missing interface is an error, never a default-route dial.
func applyBindToDevice(d *net.Dialer, iface string) {
	if iface == "" {
		return
	}
	d.Control = func(network, address string, c syscall.RawConn) error {
		selected, err := net.InterfaceByName(iface)
		if err != nil {
			return err
		}
		level, index := windows.IPPROTO_IP, selected.Index
		if network == "tcp6" || network == "udp6" {
			level = windows.IPPROTO_IPV6
		} else {
			// Windows IP_UNICAST_IF expects the IPv4 index in network order.
			var bytes [4]byte
			binary.BigEndian.PutUint32(bytes[:], uint32(index))
			index = int(binary.NativeEndian.Uint32(bytes[:]))
		}
		var sockErr error
		err = c.Control(func(fd uintptr) {
			// IP_UNICAST_IF and IPV6_UNICAST_IF both use option 31.
			sockErr = windows.SetsockoptInt(windows.Handle(fd), level, 31, index)
		})
		if err != nil {
			return err
		}
		return sockErr
	}
}
