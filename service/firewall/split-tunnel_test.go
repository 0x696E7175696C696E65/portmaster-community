package firewall

import (
	"testing"

	"github.com/safing/portmaster/service/network"
	"github.com/safing/portmaster/service/network/packet"
)

func TestSelectedRoutingCannotAcceptUnsupportedProtocol(t *testing.T) {
	// Once routing is selected, no IP protocol outside TCP/UDP may retain a
	// direct accept verdict. Cover the full byte range, including ICMP/ICMPv6.
	for protocol := 0; protocol <= 255; protocol++ {
		conn := &network.Connection{IPProtocol: packet.IPProtocol(protocol), Verdict: network.VerdictAccept}
		supported := enforceSplitTunnelProtocol(conn)
		if conn.IPProtocol == packet.TCP || conn.IPProtocol == packet.UDP {
			if !supported || conn.Verdict != network.VerdictAccept {
				t.Fatalf("supported protocol %d was denied", protocol)
			}
			continue
		}
		if supported || conn.Verdict != network.VerdictFailed || conn.Reason.Msg == "" {
			t.Fatalf("unsupported selected protocol %d retained direct access: %v", protocol, conn.Verdict)
		}
	}
}
