package updates

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDownloadLimitsAdvertisedAndStreamedBytes(t *testing.T) {
	for _, chunked := range []bool{false, true} {
		t.Run(map[bool]string{false: "content-length", true: "chunked"}[chunked], func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if chunked {
					w.(http.Flusher).Flush()
				}
				_, _ = w.Write([]byte("123456789"))
			}))
			defer server.Close()
			d := NewDownloader(nil, nil)
			d.httpClient.Transport = server.Client().Transport
			if data, err := d.downloadData(context.Background(), server.URL, 8); err == nil || data != nil {
				t.Fatal("oversized unsigned download accepted")
			}
			if data, err := d.downloadData(context.Background(), server.URL, 9); err != nil || string(data) != "123456789" {
				t.Fatalf("exact limit download rejected: %v", err)
			}
		})
	}
}

func TestDownloadRejectsHTTPAndRedirectDowngrade(t *testing.T) {
	d := NewDownloader(nil, nil)
	if _, err := d.downloadData(context.Background(), "http://127.0.0.1/untrusted", 16); err == nil {
		t.Fatal("HTTP download accepted")
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1/untrusted", http.StatusFound)
	}))
	defer server.Close()
	d.httpClient.Transport = server.Client().Transport
	if _, err := d.downloadData(context.Background(), server.URL, 16); err == nil || !strings.Contains(err.Error(), "must use HTTPS") {
		t.Fatalf("HTTP redirect not rejected by redirect policy: %v", err)
	}
}

func TestBoundedReadRejectsInsteadOfTruncating(t *testing.T) {
	if data, err := readBounded(strings.NewReader("123456789"), 8); err == nil || data != nil {
		t.Fatal("oversized content silently truncated")
	}
	if data, err := readBounded(strings.NewReader("12345678"), 8); err != nil || string(data) != "12345678" {
		t.Fatal("content at the limit rejected")
	}
}

func TestGzipRejectsCorruptChecksum(t *testing.T) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, _ = writer.Write([]byte("valid data"))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	compressed := append([]byte(nil), buf.Bytes()...)
	if data, err := Decompress("gz", compressed); err != nil || string(data) != "valid data" {
		t.Fatal("valid gzip content rejected")
	}
	compressed[len(compressed)-8] ^= 0xff
	if _, err := Decompress("gz", compressed); err == nil {
		t.Fatal("gzip with an invalid checksum accepted")
	}
}
