//go:build linux || windows

package proxy

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestMissingInterfaceDoesNotUseDefaultRoute(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	dialer := &net.Dialer{Timeout: time.Second}
	applyBindToDevice(dialer, "missing-wg-test")
	conn, err := dialer.DialContext(context.Background(), "tcp4", listener.Addr().String())
	if conn != nil {
		conn.Close()
	}
	if err == nil {
		t.Fatal("missing interface used the default route")
	}
	_ = listener.(*net.TCPListener).SetDeadline(time.Now().Add(50 * time.Millisecond))
	if conn, err := listener.Accept(); err == nil {
		conn.Close()
		t.Fatal("destination received direct traffic")
	}
}
