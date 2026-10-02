package proxy

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

func TestParseTorEndpoint(t *testing.T) {
	for _, selector := range []string{"tor", "tor://127.0.0.1:9150", "tor://[::1]:9050"} {
		if endpoint, tor, err := ParseTorEndpoint(selector); err != nil || !tor || endpoint == "" {
			t.Fatalf("%s: %q %v %v", selector, endpoint, tor, err)
		}
	}
	for _, selector := range []string{"tor:invalid", "tor://example.org:9050", "tor://192.0.2.1:9050", "tor://127.0.0.1:0", "tor://127.0.0.1:65536", "tor://127.0.0.1:abc"} {
		if _, _, err := ParseTorEndpoint(selector); err == nil {
			t.Fatalf("accepted %s", selector)
		}
	}
	if _, tor, err := ParseTorEndpoint("wg0"); tor || err != nil {
		t.Fatal("WireGuard interface interpreted as Tor")
	}
}

func TestTorSOCKSRouting(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	result := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			result <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
		greeting := make([]byte, 3)
		if _, err = io.ReadFull(conn, greeting); err != nil {
			result <- err
			return
		}
		if greeting[0] != 5 || greeting[1] != 1 || greeting[2] != 0 {
			result <- io.ErrUnexpectedEOF
			return
		}
		_, _ = conn.Write([]byte{5, 0})
		request := make([]byte, 10)
		if _, err = io.ReadFull(conn, request); err != nil {
			result <- err
			return
		}
		if request[1] != 1 || request[3] != 1 || !net.IP(request[4:8]).Equal(net.ParseIP("192.0.2.1")) || binary.BigEndian.Uint16(request[8:]) != 443 {
			result <- io.ErrUnexpectedEOF
			return
		}
		_, _ = conn.Write([]byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 1})
		payload := make([]byte, 4)
		_, err = io.ReadFull(conn, payload)
		if err == nil {
			_, err = conn.Write(payload)
		}
		result <- err
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := dialUpstream(ctx, &net.Dialer{Timeout: time.Second}, "tcp4", "192.0.2.1:443", &LocalBinding{SOCKSProxy: listener.Addr().String()})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	_, _ = conn.Write([]byte("test"))
	buffer := make([]byte, 4)
	if _, err = io.ReadFull(conn, buffer); err != nil || string(buffer) != "test" {
		t.Fatalf("echo: %q %v", buffer, err)
	}
	if err = <-result; err != nil {
		t.Fatal(err)
	}
}

func TestTorFailureNeverDialsDestination(t *testing.T) {
	destination, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer destination.Close()
	// Closed local port represents an unavailable Tor daemon.
	socks, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	endpoint := socks.Addr().String()
	_ = socks.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := dialUpstream(ctx, &net.Dialer{Timeout: time.Second}, "tcp", destination.Addr().String(), &LocalBinding{SOCKSProxy: endpoint})
	if conn != nil {
		conn.Close()
	}
	if err == nil {
		t.Fatal("Tor failure allowed a direct connection")
	}
	_ = destination.(*net.TCPListener).SetDeadline(time.Now().Add(50 * time.Millisecond))
	if conn, err := destination.Accept(); err == nil {
		conn.Close()
		t.Fatal("destination was dialed directly")
	}
}
