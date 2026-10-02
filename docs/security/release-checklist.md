# Release checklist

This checklist is a gate for a future stable release, not evidence of completion. Record the commit, reviewer, date, evidence paths, findings, and explicit decisions for each item.

- [ ] Requirements and threat model reviewed; security-sensitive changes independently approved.
- [ ] Main branch protection, required checks, access controls, MFA, isolated runner and least-privilege release credentials verified on the Git server.
- [ ] Clean source commit and lockfiles recorded; toolchain versions and build commands captured.
- [ ] Targeted gate passed; relevant integration tests completed, including DNS, IPv4/IPv6, UDP, proxy failure, WireGuard tunnel-stop, driver restart and persistence.
- [ ] npm, Go, Rust and native component vulnerability reports triaged; reachable critical/high findings fixed. Any other exception has severity, owner, rationale, mitigation, review and expiry dates.
- [ ] Secrets scan and privileged API review completed. No keys or unredacted traffic logs in source or artifacts.
- [ ] SPDX or CycloneDX SBOM generated and validated for the distributed components, including DLLs, driver and frontend. A dependency list alone is not a complete SBOM.
- [ ] Signed/authenticated release manifest or platform signing established, with independently published trusted verification instructions. Hashes alone do not prove publisher identity.
- [ ] Signed driver provenance and compatibility checked; all third-party notices and GPL source obligations preserved.
- [ ] Upgrade preserves settings/history and supports verified rollback; uninstall behavior tested.
- [ ] Source, package, hashes, SBOM, scan logs, reviews, release notes and build provenance archived durably.
- [ ] Maintainer decides whether private vulnerability reporting must be enabled before distribution; its current deferral is prominently documented.
- [ ] Release notes explain actual protections, default settings, known risks and support status; no unsupported NIST compliance or leak-protection claim.

After release, review advisories and reports, link affected releases, prioritize fixes by exploitability, publish mitigations, and record root cause and regression coverage.
