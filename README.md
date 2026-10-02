# Portmaster Community — bytefrxst fork

An independent open-source fork of [Safing Portmaster](https://github.com/safing/portmaster), with account-free local features and user-controlled privacy routing.

Repository: https://git.frxst.org/bytefrxst/portmaster

## Implemented changes

- Local network history and per-app bandwidth tracking work without a Safing account or subscription. History still follows your global/per-app recording preferences and retention settings.
- The desktop reads a local Community capability profile. Dashboard subscription controls, login forms, and the SPN sales page are replaced with community information and routing instructions.
- The integrated desktop client disables paid SPN startup, rejects enabling it, removes login/logout API endpoints, and stops periodic account/token refresh. The local capability profile does not grant hosted SPN or Safing priority support.
- Tor TCP routing uses a local SOCKS5 daemon, with IPv4/IPv6 loopback endpoint validation and no direct fallback on proxy failure. Tor-selected UDP is rejected.
- WireGuard routing reuses interface-based TCP/UDP routing. Linux binds the socket to the selected device; Windows additionally pins the outgoing interface using IP_UNICAST_IF/IPV6_UNICAST_IF.
- Selected traffic is blocked if the routing module is unavailable. Missing selected interfaces and failed Tor connections fail closed for new routed connections.
- Safing binary update feeds are removed and automatic software updates default to off, preventing default feed downloads from replacing this fork. Intelligence-data updates remain upstream.

This is a development source fork. It has no release installer yet. Legacy SPN internals remain for upstream compatibility and standalone hub tooling; they are disabled in the integrated client. This initial version does not remove every legacy SPN configuration label or Safing asset.

## Tor setup

1. Install and run a local Tor daemon with a SOCKS listener on 127.0.0.1:9050. Another port is supported.
2. Enable **Split Tunnel Module** globally.
3. For the selected app, enable **Use Split Tunnel** and set **Network Interface** to `tor`. Use `tor://127.0.0.1:9150` or `tor://[::1]:9050` for another listener.
4. Keep routing disabled for the Tor daemon itself to avoid routing its own relay connections back through Tor.
5. Leave the app's Split Tunnel rules empty to route all otherwise allowed TCP traffic. A deny rule in this routing policy deliberately excludes matching traffic from the tunnel; firewall blocking rules are separate.

Tor does not support SOCKS UDP association, so UDP traffic selected for Tor is blocked. DNS resolution still uses Portmaster's configured resolver independently of Tor. This does not provide .onion hostname resolution, per-app circuit isolation, or Tor Browser fingerprint protection. See the [Tor SOCKS specification](https://spec.torproject.org/socks-extensions.html).

## WireGuard setup

1. Install WireGuard and configure/start your own tunnel using your chosen peer, keys, routes, and AllowedIPs.
2. Enable **Split Tunnel Module**, then **Use Split Tunnel** for the selected app.
3. Set **Network Interface** to the running tunnel interface name (for example `wg0`) or its assigned local IP.
4. Exclude the WireGuard process from app routing. Ensure the tunnel supports every destination and IP family you intend to route; an absent family is rejected.
5. Configure DNS separately. Portmaster does not create WireGuard keys, provision peers, or operate a VPN service.

See [WireGuard's quick start](https://www.wireguard.com/quickstart/). Selected new outbound connections using unsupported protocols are blocked instead of using the default route. Existing connections, DNS, localhost traffic, inbound traffic, and deliberately excluded traffic are outside the new-connection routing policy. The full desktop/kernel path still needs live tunnel and disconnect testing before a production release.

## Build and validation

Use the patched Go 1.26.8 toolchain declared in `go.mod`. Build the Windows core with:

```powershell
go build -o portmaster-core.exe ./cmds/portmaster-core
go test -short ./service/splittun/proxy ./spn/access ./spn/captain ./service/network ./service/firewall ./service/profile ./service/splittun ./service/configure ./service/core
```

The short flag skips upstream tests that require a live paid SPN account. Added tests cover account-free capabilities, Tor endpoint validation, SOCKS negotiation and data forwarding, UDP refusal, failed Tor connections without direct fallback, and unavailable interface binding.

For the Angular desktop:

```powershell
cd desktop/angular
npm ci --ignore-scripts
npm run build-libs:dev
npx ng build --configuration development
```

Building the core does not install its kernel driver or system service. See [the preserved upstream README](README.upstream.md) and platform packaging sources for those requirements.

## Provenance and license

Based on Safing's development commit 13a86a43cc6cee592395fcddc8387178f290f144, with upstream Git history retained. Fork changes are dated 2026-10-02. The Go module path remains github.com/safing/portmaster to keep internal imports compatible.

The upstream GPL-3.0 license, copyright notices, and bundled asset licenses are preserved. This fork is not an official Safing release. See [LICENSE](LICENSE).

## Community help and secure development

See [Getting started](docs/community/getting-started.md), [Routing and settings](docs/community/routing.md), [Contributing](CONTRIBUTING.md), and [Security policy](SECURITY.md). The [SSDF process and gap map](docs/security/ssdf.md) records our development requirements and outstanding evidence. This experimental fork does not claim NIST certification or comprehensive security assurance.

The [2026-10-02 security audit](docs/security/audit-2026-10-02.md) records fixes, validation, and remaining limitations. Security checks run from `.github/workflows/community-security.yml` and `.gitea/workflows/community-security.yml`, with read-only permissions and pinned actions/scanners. Gitea requires a configured isolated Windows runner; server-side required checks and branch protection still need verification. Upstream publishing workflows are archived outside the active workflow directory.
