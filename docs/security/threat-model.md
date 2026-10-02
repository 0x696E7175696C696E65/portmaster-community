# Initial threat model

Scope: Windows community core service, native desktop and Angular UI; per-app Tor SOCKS forwarding and WireGuard interface binding. Linux routing is inherited and modified but not integration-tested here. Assets include firewall decisions, browsing metadata/history, API credentials, configuration, release integrity, and host privileges. The service runs as LocalSystem and includes a kernel driver.

Boundaries: untrusted network to firewall/proxy; local applications to API; desktop to privileged service; build tools/dependencies to release; local proxy/VPN to selected application; maintainer Git credentials to repository. Loopback limits exposure but is not authentication. A malicious local process or compromised UI must be considered.

| Requirement | Threat and expected behavior | Evidence / outstanding work |
| --- | --- | --- |
| SEC-01 | No public arbitrary-URL binary update route; no upstream binary replacement through normal update actions | Community API registration test; local installer bypasses remote binary feeds |
| SEC-02 | Selected Tor TCP must fail when the proxy cannot connect; reject selected UDP | SOCKS negotiation, failure/no-fallback and UDP tests; real tunnel test pending |
| SEC-03 | Accept only literal loopback Tor endpoint and a valid port; avoid remote proxy and ambiguous host resolution | Endpoint parser tests |
| SEC-04 | Bind WireGuard selection to intended interface; reject absent interface; do not claim peer health | Missing-interface tests; live interface/tunnel-stop IPv4/IPv6 tests pending |
| SEC-05 | API state changes require authorization; loopback is not a browser trust boundary | Diagnostics require user access; HTTP/WebSocket origin and loopback Host checks, streamed request bounds and credential redaction; broader endpoint review remains ongoing |
| SEC-06 | Preserve data during upgrade; validate payload integrity; authenticate release source | SHA-256 manifest; driver signature; rollback upgrade; fork publisher signing pending |
| SEC-07 | No automatic report uploads; redact sensitive diagnostics and private keys | Help opens repository; old support ticket routes redirect |
| SEC-08 | Preserve signed intelligence verification independently from community binaries | Original signature verification retained; downloaded index verified; ongoing review required |
| SEC-09 | UI reports unavailable/stale data and makes no fabricated tunnel status claim | Dashboard retry status and setup instructions; polling visual check |
| SEC-10 | Untrusted profile/API metadata cannot become executable markup | Text bindings, Angular sanitization, validated country flags, production CSP and sanitizer guard; browser metadata fixture |
| SEC-11 | Native IPC cannot be used for arbitrary network destinations | Exact local HTTP/database WebSocket URLs, no redirects/proxies, bounded sessions/messages/time, explicit command ACL and trusted navigation |
| SEC-12 | Update permission/rollback failure must not report success | ACL error propagation, staged-copy rejection, purge preparation failure and aggregate rollback error tests |
| SEC-13 | Kernel input/output must be sized, initialized and used within IRP lifetime | Bounds-checked protocol and IRP source fixes; shipping signed driver still requires rebuilt validation |

Additional risks: WireGuard interface binding alone is not a peer-aware kill switch; Tor DNS is separate, onion resolution and circuit isolation are absent; history is sensitive local data; native shells, asset parsing, dependency vulnerabilities, LocalSystem API handling, driver loading and crash recovery require further review. End-to-end testing and a full privileged API audit remain release requirements.

Review this model whenever trust boundaries, listeners, update paths, tunnel selection, dependencies, installer permissions, or persisted data change.
