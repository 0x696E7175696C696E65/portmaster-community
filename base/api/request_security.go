package api

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// allowedRequestHost prevents DNS rebinding when the API listens on loopback.
// Non-loopback listeners retain the configured API key access behavior.
func allowedRequestHost(host, listenAddress string) bool {
	listenHost, listenPort, err := net.SplitHostPort(listenAddress)
	if err != nil || !isLoopbackHost(listenHost) {
		return true
	}
	requestHost, requestPort, err := net.SplitHostPort(host)
	if err != nil {
		// An omitted port denotes HTTP port 80, not the API listener's port.
		requestHost, requestPort = host, "80"
	}
	return requestPort == listenPort && isLoopbackHost(requestHost)
}

// allowedRequestOrigin is shared by HTTP and WebSocket handlers. A serialized
// browser origin contains only a scheme and authority, with an exact port.
// Native clients without an Origin still pass through normal authentication.
func allowedRequestOrigin(r *http.Request) bool {
	origins := r.Header.Values("Origin")
	if len(origins) == 0 {
		return true
	}
	if len(origins) != 1 || origins[0] == "" {
		return false
	}
	origin, err := url.Parse(origins[0])
	if err != nil || origin.Host == "" || origin.User != nil ||
		origin.Path != "" || origin.ForceQuery || origin.RawQuery != "" ||
		strings.Contains(origins[0], "#") ||
		(origin.Scheme != "http" && origin.Scheme != "https") {
		return false
	}
	if devMode != nil && devMode() && isLoopbackHost(origin.Hostname()) {
		return true
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if origin.Scheme != scheme {
		return false
	}
	requestURL, err := url.Parse(scheme + "://" + r.Host)
	if err != nil || requestURL.Host == "" || requestURL.User != nil || requestURL.Path != "" {
		return false
	}
	defaultPort := "80"
	if scheme == "https" {
		defaultPort = "443"
	}
	originPort, requestPort := origin.Port(), requestURL.Port()
	if originPort == "" {
		originPort = defaultPort
	}
	if requestPort == "" {
		requestPort = defaultPort
	}
	return strings.EqualFold(origin.Hostname(), requestURL.Hostname()) && originPort == requestPort
}
