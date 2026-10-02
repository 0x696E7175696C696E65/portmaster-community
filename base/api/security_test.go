package api

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortUnknownAPIKeysDoNotPanic(t *testing.T) {
	for _, value := range []string{"Bearer ", "Bearer a", "Bearer ab", "Bearer abc", "Basic Og=="} {
		r := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
		r.Header.Set("Authorization", value)
		if token := checkAPIKey(r); token != nil {
			t.Fatalf("malformed or unknown credential %q was accepted", value)
		}
	}
}

type zeroReader struct{}

func (zeroReader) Read(data []byte) (int, error) {
	clear(data)
	return len(data), nil
}

func TestRequestBodyLimitsActualBytes(t *testing.T) {
	for _, contentLength := range []int64{-1, 0, maxAPIRequestBodySize + 1} {
		r := httptest.NewRequest(http.MethodPost, "http://localhost/test", io.LimitReader(zeroReader{}, maxAPIRequestBodySize+1))
		r.ContentLength = contentLength
		w := httptest.NewRecorder()
		if body, ok := readBody(w, r); ok || body != nil || w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("oversized body with Content-Length %d: accepted=%t, status=%d", contentLength, ok, w.Code)
		}
	}
	r := httptest.NewRequest(http.MethodPost, "http://localhost/test", strings.NewReader("small request"))
	r.ContentLength = -1
	if data, ok := readBody(httptest.NewRecorder(), r); !ok || string(data) != "small request" {
		t.Fatal("valid streamed request was not preserved")
	}
}

func TestStrictRequestOrigin(t *testing.T) {
	for _, test := range []struct {
		host, origin string
		tls          bool
		allowed      bool
	}{
		{"127.0.0.1:817", "http://127.0.0.1:817", false, true},
		{"localhost:817", "http://localhost:817", false, true},
		{"localhost", "http://localhost:80", false, true},
		{"localhost:80", "http://localhost", false, true},
		{"localhost:817", "", false, true},
		{"localhost", "http://localhost:9999", false, false},
		{"localhost:817", "http://localhost:9999", false, false},
		{"localhost:817", "https://localhost:817", false, false},
		{"localhost:817", "http://localhost:817", true, false},
		{"localhost:817", "https://localhost:817", true, true},
		{"localhost:817", "http://user@localhost:817", false, false},
		{"localhost:817", "http://localhost:817/path", false, false},
		{"localhost:817", "http://localhost:817?query", false, false},
		{"localhost:817", "http://localhost:817?", false, false},
		{"localhost:817", "http://localhost:817#", false, false},
		{"localhost:817", "chrome-extension://untrusted", false, false},
		{"localhost:817", "null", false, false},
		{"localhost:817", "http://attacker.example:817", false, false},
	} {
		t.Run(test.host+"/"+test.origin, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "http://"+test.host+"/test", nil)
			if test.origin != "" {
				r.Header.Set("Origin", test.origin)
			}
			if test.tls {
				r.TLS = &tls.ConnectionState{}
			}
			if got := allowedRequestOrigin(r); got != test.allowed {
				t.Fatalf("allowed=%t, want %t", got, test.allowed)
			}
		})
	}
	r := httptest.NewRequest(http.MethodGet, "http://localhost:817/test", nil)
	r.Header.Add("Origin", "http://localhost:817")
	r.Header.Add("Origin", "http://attacker.example")
	if allowedRequestOrigin(r) {
		t.Fatal("multiple Origin headers were accepted")
	}
}

func TestLoopbackListenerRejectsDNSRebindingHost(t *testing.T) {
	for _, host := range []string{"attacker.example:817", "127.0.0.1.attacker.example:817", "localhost:9999", "localhost", "127.0.0.1:817@attacker.example"} {
		if allowedRequestHost(host, "127.0.0.1:817") {
			t.Fatalf("untrusted authority %q accepted on loopback", host)
		}
	}
	for _, host := range []string{"localhost:817", "127.0.0.1:817", "[::1]:817"} {
		if !allowedRequestHost(host, "127.0.0.1:817") {
			t.Fatalf("loopback client %q rejected", host)
		}
	}
	if !allowedRequestHost("api.example:817", "0.0.0.0:817") {
		t.Fatal("intentional non-loopback listener's host rejected")
	}
}

func TestDebugEndpointsRequireAuthenticatedUser(t *testing.T) {
	previousToken := *testToken
	defer func() { *testToken = previousToken }()
	*testToken = AuthToken{Read: PermitAnyone, Write: PermitAnyone}
	handler := &mainHandler{mux: mainMux}
	for _, path := range []string{"debug/stack", "debug/stack/print", "debug/cpu", "debug/heap", "debug/allocs", "debug/info"} {
		r := httptest.NewRequest(http.MethodGet, "http://localhost"+apiV1Path+path, nil)
		w := httptest.NewRecorder()
		if err := handler.handle(w, r); err != nil {
			t.Fatal(err)
		}
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("unauthenticated %s returned %d", path, w.Code)
		}
	}
	for _, path := range []string{"ping", "ready"} {
		r := httptest.NewRequest(http.MethodGet, "http://localhost"+apiV1Path+path, nil)
		w := httptest.NewRecorder()
		if err := handler.handle(w, r); err != nil {
			t.Fatal(err)
		}
		if w.Code != http.StatusOK {
			t.Fatalf("public health endpoint %s returned %d", path, w.Code)
		}
	}
}

func TestCPUProfilingRejectsUnboundedDuration(t *testing.T) {
	for _, duration := range []string{"0s", "-1s", "61s", "87600h"} {
		r := httptest.NewRequest(http.MethodGet, "http://localhost/debug/cpu?duration="+duration, nil)
		_, err := handleCPUProfile(&Request{Request: r})
		var status HTTPStatusProvider
		if !errors.As(err, &status) || status.HTTPStatus() != http.StatusBadRequest {
			t.Fatalf("duration %s: expected rejection before starting profiler, got %v", duration, err)
		}
	}
}

func TestProductionSecurityPolicyIsAppliedToResponses(t *testing.T) {
	previousDevMode := devMode
	defer func() { devMode = previousDevMode }()
	handler := &mainHandler{mux: mainMux}
	for _, development := range []bool{false, true} {
		devMode = func() bool { return development }
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "http://localhost"+apiV1Path+"ping", nil)
		if err := handler.handle(w, r); err != nil {
			t.Fatal(err)
		}
		policy := w.Header().Get("Content-Security-Policy")
		if development {
			if policy != "" {
				t.Fatal("development tooling unexpectedly received production CSP")
			}
			continue
		}
		for _, required := range []string{"script-src 'self'", "base-uri 'self'", "form-action 'self'", "object-src 'none'", "frame-ancestors 'none'", "connect-src 'self' ipc: http://ipc.localhost", "style-src 'self' 'unsafe-inline'"} {
			if !strings.Contains(policy, required) {
				t.Fatalf("production response missing %s", required)
			}
		}
		if strings.Contains(policy, "safing.io") || strings.Contains(policy, "unsafe-eval") || strings.Contains(policy, "script-src 'self' 'unsafe-inline'") {
			t.Fatal("production CSP allows external connections or executable inline/eval content")
		}
	}
}
