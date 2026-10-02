# Getting started and FAQ

The bytefrxst Portmaster community fork provides a local firewall, connection history, bandwidth visibility, and per-application routing without a subscription. It is experimental. Source and reports belong in [this repository](https://git.frxst.org/bytefrxst/portmaster).

## Installation

The Windows build needs the core service, native desktop, UI archive, signed compatible kernel driver, DLLs, and intelligence data. A standalone `go build` is not an installer. Use a maintainer-provided package with a manifest and verify hashes and provenance before running it as administrator. No public signed community release is currently declared. Development packages may be unsigned; a checksum does not authenticate the publisher.

The current local installation uses `C:\Program Files\Portmaster Community` and stores configuration, history, intelligence, and logs in `C:\ProgramData\PortmasterCommunity`. Back up data before updates. Do not run another Portmaster installation alongside this service. Uninstalling should preserve user data unless you deliberately remove it.

## Common questions

- **Do I need an account?** Local history and bandwidth are available without one. Hosted SPN is disabled in this fork.
- **Why is bandwidth empty?** The dashboard shows recent live connections. Wait for application traffic; check service availability. Internal connections are excluded from the top-consumer list.
- **What does Active Apps mean?** Applications with currently active allowed connections in the live database.
- **Can it configure Tor or WireGuard for me?** No. Run your own local Tor daemon or WireGuard tunnel, then configure [per-application routing](routing.md).
- **Does enabling routing prove I am protected?** No. Verify the selected application, tunnel availability, DNS path, and tunnel-stop behavior. See the routing limitations.
- **Where do updates come from?** Upstream binary updates are disabled to preserve the fork. Intelligence updates remain separate. Apply a verified community package manually.
- **Where do I get help?** Search this fork's issues. Follow [the reporting policy](../../SECURITY.md) and redact logs before posting.

For development builds and required checks, see [CONTRIBUTING](../../CONTRIBUTING.md).
