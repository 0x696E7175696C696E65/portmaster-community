# Initial risk register

Owner for all entries: bytefrxst. Status recorded 2026-10-02. These items remain open until evidence is added; documentation is not risk acceptance.

| ID | Finding | State and next action |
| --- | --- | --- |
| R-001 | Legacy arbitrary-URL binary upgrade API required no caller permission | Route removed; registration regression passes; installed endpoint absence must be verified after upgrade |
| R-002 | Frontend inherits an aging dependency graph | npm audit: 90 findings (4 critical, 45 high, 32 moderate, 9 low). Critical paths include inherited build/test dependencies form-data/request, piscina, and tar; runtime high findings remain. Assess reachability and migrate the Angular toolchain; no stable release approval |
| R-003 | WireGuard tunnel health and DNS leak behavior are unverified | Run live tunnel-stop, interface, DNS and IPv4/IPv6 matrix before protection claims |
| R-004 | Fork builds lack authenticated publisher signing and a complete standard SBOM | Establish trusted signing/provenance and SPDX/CycloneDX inventory before stable release |
| R-005 | Independent review and Git server required-check enforcement are absent or unverified | Assign reviewer and verify protected-branch/runner configuration |
| R-006 | Private reporting channel deferred by maintainer | Keep deferral visible; do not direct sensitive details to public issues |
| R-007 | Full privileged API, native shell, driver and Go/Rust dependency reviews incomplete | Expand source/scanner coverage and triage results |

For each new finding record: date, reporter, affected versions, severity and exploitability, root cause, owner, mitigation, fix commit, regression evidence, review date and any exception expiry. Do not include sensitive exploit details in a public register before coordinated disclosure.
