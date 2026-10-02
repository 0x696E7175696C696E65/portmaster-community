package proxy

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	xproxy "golang.org/x/net/proxy"
)

// ParseTorEndpoint recognizes the routing selector and restricts SOCKS to a
// literal loopback address. Hostnames are deliberately not resolved here.
func ParseTorEndpoint(selector string) (endpoint string, isTor bool, err error) {
	if selector == "tor" {
		return "127.0.0.1:9050", true, nil
	}
	if !strings.HasPrefix(selector, "tor:") {
		return "", false, nil
	}
	if !strings.HasPrefix(selector, "tor://") {
		return "", true, fmt.Errorf("use tor or tor://loopback-IP:port")
	}
	endpoint = strings.TrimPrefix(selector, "tor://")
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", true, fmt.Errorf("invalid Tor SOCKS endpoint: %w", err)
	}
	ip := net.ParseIP(host)
	n, portErr := strconv.Atoi(port)
	if ip == nil || !ip.IsLoopback() || portErr != nil || n < 1 || n > 65535 {
		return "", true, fmt.Errorf("Tor SOCKS requires a literal loopback IP and port 1-65535")
	}
	return endpoint, true, nil
}

func dialUpstream(ctx context.Context, dialer *net.Dialer, network, address string, binding *LocalBinding) (net.Conn, error) {
	if binding == nil || binding.SOCKSProxy == "" {
		return dialer.DialContext(ctx, network, address)
	}
	endpoint, _, err := ParseTorEndpoint("tor://" + binding.SOCKSProxy)
	if err != nil {
		return nil, err
	}
	// A separate unbound dialer connects to the local Tor daemon. Never fall
	// back to a direct connection when SOCKS setup or negotiation fails.
	forward := &net.Dialer{Timeout: dialer.Timeout}
	// Dialer.Timeout limits the local TCP connect, but the SOCKS handshake
	// also needs a deadline. A reachable daemon can accept and then stall.
	if dialer.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, dialer.Timeout)
		defer cancel()
	}
	socks, err := xproxy.SOCKS5("tcp", endpoint, nil, forward)
	if err != nil {
		return nil, err
	}
	return socks.(xproxy.ContextDialer).DialContext(ctx, "tcp", address)
}
