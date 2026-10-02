# Portmaster Community

**Application firewall, network visibility, and per-app privacy routing.**

Portmaster Community is an independent open-source fork of [Safing Portmaster](https://github.com/safing/portmaster). It provides account-free local history and bandwidth visibility, with routing through your own Tor daemon or WireGuard tunnel.

[Getting started](docs/community/getting-started.md) · [Routing guide](docs/community/routing.md) · [Contributing](CONTRIBUTING.md) · [Security](SECURITY.md) · [Issues](https://github.com/0x696E7175696C696E65/portmaster-community/issues)

> **Project status: experimental.** This is a development source fork; no public signed community release is declared. Driver validation, full Linux native integration, and live tunnel testing remain release requirements. See the [release checklist](docs/security/release-checklist.md).

## Features

| Capability | What it provides |
| --- | --- |
| Application firewall | Per-application connection controls and visibility into allowed and blocked traffic. |
| Secure DNS and privacy filtering | The inherited Portmaster resolver and filtering capabilities. |
| Network history | Local connection history without an account or subscription, subject to your recording and retention preferences. |
| Bandwidth visibility | Per-application bandwidth tracking and dashboard activity without a subscription. |
| Tor routing | Selected TCP connections through a local SOCKS5 daemon, with loopback endpoint validation and no direct fallback on proxy failure. |
| WireGuard routing | Selected TCP/UDP connections through an existing tunnel interface, using interface binding on Windows and Linux. |

### Community edition behavior

Subscription controls and account screens are replaced with community information and routing guidance. Hosted SPN is disabled in the integrated client; Tor and WireGuard use infrastructure you configure and control.

Upstream binary update feeds are removed, and automatic software updates default to off. Intelligence-data updates remain separate and continue to use upstream sources. Apply community software updates through a verified maintainer package.

Legacy SPN internals, some configuration labels, and upstream assets remain for compatibility and standalone hub tooling. This fork does not provide Safing's hosted network or priority support.

## Getting started

The Windows application requires the core service, desktop client, UI assets, a signed compatible kernel driver, DLLs, and intelligence data. Building one component does not install the complete application.

Start with the [installation and FAQ guide](docs/community/getting-started.md). Back up configuration and history before upgrading, and verify package hashes and publisher provenance before running an installer as administrator.

### Route an application through Tor

1. Run a local Tor daemon with a SOCKS listener, such as `127.0.0.1:9050`.
2. Enable **Split Tunnel Module** globally.
3. Enable **Use Split Tunnel** for the application, then set **Network Interface** to `tor`.
4. For another listener, use `tor://127.0.0.1:9150` or `tor://[::1]:9050`.
5. Keep routing disabled for the Tor daemon itself. Leave the application's Split Tunnel rules empty to select all otherwise allowed TCP traffic.

Tor-selected UDP is blocked. DNS uses Portmaster's configured resolver separately; this integration does not provide `.onion` resolution, per-app circuit isolation, or Tor Browser fingerprint protection.

### Route an application through WireGuard

1. Configure and start your own WireGuard tunnel, including its peers, keys, routes, and `AllowedIPs`.
2. Enable **Split Tunnel Module** globally and **Use Split Tunnel** for the application.
3. Set **Network Interface** to the running tunnel's interface name, such as `wg0`, or its assigned local IP.
4. Keep routing disabled for the WireGuard process itself. Configure DNS separately and ensure the tunnel supports the required destinations and IP families.

Portmaster Community selects an existing interface; it does not provision WireGuard peers or create keys.

### Routing boundaries

Selected new outbound traffic is blocked when the routing module or selected interface is unavailable. Failed Tor connections have no direct fallback, and unsupported selected protocols are blocked.

Existing connections, DNS, localhost traffic, inbound traffic, and deliberately excluded traffic are outside this new-connection routing policy. A deny rule in the **Split Tunnel** policy excludes matching traffic from the tunnel; use firewall blocking rules to block it.

See the [routing guide](docs/community/routing.md) for settings and the required DNS, IPv4/IPv6, and tunnel-stop validation before relying on a configuration.

## Build from source

### Prerequisites

| Component | Validated toolchain |
| --- | --- |
| Go core | Go 1.26.8, declared in `go.mod` |
| Angular interface | Node.js 24.20.0 and the committed npm lockfile |
| Native desktop | Rust 1.98.1, pinned in `desktop/tauri/rust-toolchain.toml`, plus platform build dependencies |

Use the committed lockfiles. Additional driver and packaging prerequisites are described in the [preserved upstream documentation](README.upstream.md) and platform packaging sources.

### Clone the repository

```powershell
git clone https://github.com/0x696E7175696C696E65/portmaster-community.git
cd portmaster-community
```

A [Git server mirror](https://git.frxst.org/bytefrxst/portmaster) is also maintained.

### Build the Windows core

```powershell
go build -o portmaster-core.exe ./cmds/portmaster-core
```

### Build the interface

```powershell
cd desktop/angular
npm ci --ignore-scripts
npm run build
npm run build-tauri
cd ../..
```

Build the main interface before the Tauri builtin interface: their outputs share `dist`, and the main build cleans that directory. Native compilation and system-service installation are separate steps.

### Run the development checks

From the repository root, after installing the frontend dependencies:

```powershell
./scripts/security-check.ps1 -GoCommand go -BuildUI
```

The gate runs targeted API, updater, permission, routing, and UI regressions, Go vet, and production builds. CI adds dependency scans, native workspace tests, and Linux checks. See [CONTRIBUTING.md](CONTRIBUTING.md) for required checks, integration prerequisites, and evidence collection.

## Security and development process

Development follows the project's [NIST SSDF process and gap map](docs/security/ssdf.md). Security-sensitive changes require documented requirements, review, regression coverage, dependency triage, and release evidence.

The [2026-10-02 security audit](docs/security/audit-2026-10-02.md) records fixes, test results, scanner limitations, and remaining work. Outstanding items are tracked in the [risk register](docs/security/risk-register.md). This project does not claim NIST certification or comprehensive security assurance.

GitHub and Gitea security workflows use pinned actions and scanners, read-only permissions, and checkout without persistent credentials. Server-side required checks, branch protection, and isolated Gitea runners still require verification.

For vulnerability reporting, follow [SECURITY.md](SECURITY.md). Private reporting is currently deferred; do not post sensitive vulnerability details, credentials, or unredacted traffic logs in public issues.

## Contributing and support

Bug reports, documentation improvements, and reviewed code contributions are welcome. Use the fork's [issues](https://github.com/0x696E7175696C696E65/portmaster-community/issues) and follow the [contribution guide](CONTRIBUTING.md).

For setup questions, consult the [FAQ](docs/community/getting-started.md) and [routing documentation](docs/community/routing.md).

## License and provenance

Licensed under [GPL-3.0](LICENSE), with upstream copyright notices and bundled asset licenses preserved. This is an independent community project and is not an official Safing release.

Based on Safing's development commit `13a86a43cc6cee592395fcddc8387178f290f144`, with upstream Git history retained. The Go module path remains `github.com/safing/portmaster` for internal import compatibility. Reviewed dependency backports retain their original licenses and provenance alongside the vendored source.
