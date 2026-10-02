# Risk register

Owner for all entries: bytefrxst. Status recorded 2026-10-02. These items remain open until evidence is added; documentation is not risk acceptance.

| ID | Finding | State and next action |
| --- | --- | --- |
| R-001 | Legacy arbitrary-URL binary upgrade API required no caller permission | Route removed; registration regression passes; installed endpoint absence must be verified after upgrade |
| R-002 | Inherited dependency vulnerability and maintenance risks | 2026-10-02 audit: npm full 90→0 and production 23→0 after Angular 21 migration; Go core reachable 10→0; native registry vulnerability records 20→0. Eleven Rust informational warnings remain: ten unmaintained and rand's log feature absent on Windows. Cargo Audit omits local packages, so GLib's original advisory is tracked separately with a reviewed source backport, five passing optimized Linux regressions and a crashing unpatched negative control. Full Linux native integration remains a release gate. Owner bytefrxst; review by 2026-10-09 and at dependency changes |
| R-003 | WireGuard tunnel health and DNS leak behavior are unverified | Run live tunnel-stop, interface, DNS and IPv4/IPv6 matrix before protection claims |
| R-004 | Fork builds lack authenticated publisher signing and a complete standard SBOM | Establish trusted signing/provenance and SPDX/CycloneDX inventory before stable release |
| R-005 | Independent review and Git server required-check enforcement are absent or unverified | Assign reviewer and verify protected-branch/runner configuration |
| R-006 | Private reporting channel deferred by maintainer | Keep deferral visible; do not direct sensitive details to public issues |
| R-007 | Review scope remains incomplete | Bounded API/native/routing/update/dependency audit and regression fixes completed; full kernel concurrency/WFP fuzzing, memory safety and end-to-end driver/native tests still required; see audit report |
| R-008 | Driver buffer/lifetime source fixes are absent from shipped signed binary | Source parser and IRP helpers patched; protocol tests pass. Full driver compilation/WDK/Verifier/signing/deployment unverified. Administrator/SYSTEM prerequisite limits exposure but does not close the finding. No stable release approval until patched signed artifact validated |
| R-009 | Inherited OnceAgain execution-count test is scheduling-sensitive | Full base/utils run failed under concurrent build load on Windows/Linux; focused permission tests pass. Investigate deterministic concurrency coverage before declaring the full utility suite green; owner bytefrxst, review by 2026-10-09 |

For each new finding record: date, reporter, affected versions, severity and exploitability, root cause, owner, mitigation, fix commit, regression evidence, review date and any exception expiry. Do not include sensitive exploit details in a public register before coordinated disclosure.
