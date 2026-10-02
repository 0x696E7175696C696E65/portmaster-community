# Contributing

Use this fork's [issues](https://git.frxst.org/bytefrxst/portmaster/issues) and pull requests. Preserve upstream copyright, license notices, and source provenance.

Before changing code, describe the user problem, expected behavior, affected trust boundaries, and failure cases in an issue or pull request. For routing, API authorization, driver, update, installation, or storage changes, update the threat model and link the relevant security requirement. Never commit keys, tokens, credentials, personal traffic captures, or unredacted debug logs.

Use Go 1.26.8 as specified by `go.mod`, Node 24.20.0 for Angular 21, and the validated Rust 1.98.1 desktop toolchain. Install JavaScript dependencies using `npm ci --ignore-scripts`; review and run any necessary dependency build scripts individually. Use lockfiles and inspect dependency, license, and vulnerability changes. Do not use `npm audit fix --force` without reviewing behavior and compatibility changes.

Run the portable gate from the repository root:

```powershell
./scripts/security-check.ps1 -GoCommand go -BuildUI
```

`-BuildUI` assumes `npm ci --ignore-scripts` has completed in `desktop/angular`. The gate runs API/update/routing tests, Windows permission regressions, Go vet, a core build, chart/navigation/security regressions, and Angular production builds. It fails when any requested check fails. Use `-EvidenceDirectory` to preserve logs outside the working tree. CI additionally runs pinned `govulncheck`, `npm audit`, native workspace tests, and `cargo audit`. Native production tests require the Tauri frontend to be built after the main Angular build, which otherwise removes that output directory. Preserve scanner exit codes and triage informational warnings as well as vulnerability records; do not silently suppress findings. Two privileged updater integration tests skip on unelevated Windows and must also run on Linux or elevated Windows before release.

Review changes before merging. The maintainer, bytefrxst, owns security triage and release decisions. Security-sensitive changes require an independent reviewer; if one is unavailable, document the missing review and keep the change experimental. Protect the main branch, require passing checks, enable MFA, and use least-privilege credentials. These server settings must be verified separately; repository files cannot enable them.

A pull request should include reproduction steps, security impact, meaningful validation results, dependency changes, rollback instructions, and remaining limitations. Do not claim Tor or WireGuard leak protection without recording real tunnel-stop, DNS, IPv4/IPv6, and UDP tests on each supported platform.
