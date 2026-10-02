package splittun

import (
	"github.com/safing/portmaster/service/network"
	"github.com/safing/portmaster/service/network/packet"
	"testing"
)

func TestTorRejectsUDPBeforeQueueing(t *testing.T) {
	conn := &network.Connection{IPProtocol: packet.UDP}
	if _, err := AwaitRequest(conn, "tor"); err == nil {
		t.Fatal("UDP accepted for Tor routing")
	}
}
