package api

import (
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestDatabaseWebsocketRejectsOversizedDeclaredMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(startDatabaseWebsocketAPI))
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close() //nolint:errcheck

	// A masked binary frame declares its size before its payload. The server
	// must reject this without reading/allocating the advertised large body.
	frame := make([]byte, 14)
	frame[0], frame[1] = 0x82, 0xff
	binary.BigEndian.PutUint64(frame[2:10], maxAPIRequestBodySize+1)
	if _, err := conn.UnderlyingConn().Write(frame); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseMessageTooBig) {
		t.Fatalf("oversized message was not rejected before its payload: %v", err)
	}
}

func TestDatabaseWebsocketRejectsCrossOrigin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(startDatabaseWebsocketAPI))
	defer server.Close()
	conn, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), http.Header{"Origin": {"http://attacker.example"}})
	if conn != nil {
		_ = conn.Close()
	}
	if err == nil || response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin upgrade was not denied: response=%v, error=%v", response, err)
	}
}

func TestDatabaseQueueSenderStopsOnDisconnect(t *testing.T) {
	api := &DatabaseWebsocketAPI{
		DatabaseAPI: DatabaseAPI{shutdownSignal: make(chan struct{})},
		sendQueue:   make(chan []byte, 1),
	}
	api.sendQueue <- []byte("full queue")
	returned := make(chan struct{})
	go func() {
		api.enqueue([]byte("pending response"))
		close(returned)
	}()
	close(api.shutdownSignal)
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("disconnected client left a response sender blocked")
	}
}
