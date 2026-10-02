# Secure software development lifecycle

Baseline: [NIST SP 800-218, SSDF 1.1 (final)](https://csrc.nist.gov/pubs/sp/800/218/final). This document defines this fork's process and records adoption gaps. It is not a certification, audit report, or claim that all SSDF practices are implemented. Reassess the baseline when a revision becomes final.

## Workflow and evidence

For every change: record requirements and abuse cases; assess design and dependencies; implement with least privilege and explicit failure behavior; review; run meaningful tests and scans; triage findings; assemble verified artifacts; document rollback; and monitor reports after release. Link the issue, commit, review, tool versions, results, and unresolved risks. Security-sensitive changes stay experimental until independent review and required integration tests are complete.

Owner: bytefrxst. Independent reviewer: not yet assigned. Private reporting: deferred. CI runner and required main-branch checks: not yet verified. Maintain these facts rather than marking a control complete because a file exists.

## Practice mapping

The following is a project-specific implementation map using NIST practice identifiers; consult the publication for its full task definitions.

| Practice | Project action and evidence | Current state / gap |
| --- | --- | --- |
| PO.1 | Requirements in threat model; security impact in PR template | Initial requirements documented; acceptance tracked per change |
| PO.2 | Maintainer owns triage; independent review required by contribution policy | Owner named; second reviewer and training evidence pending |
| PO.3 | Locked Go/npm/Cargo dependencies; portable check script with logs | Local gate available; scanner coverage and CI runner pending |
| PO.4 | Release checklist defines failure criteria and records exceptions | Criteria documented; no stable release approval recorded |
| PO.5 | Isolated test machine, least privilege, MFA, reviewed credentials and toolchains | Local development only; server and runner isolation unverified |
| PS.1 | Preserve Git history; protect main branch; restrict write access | History retained; access review and branch protection unverified |
| PS.2 | Artifact SHA-256 manifest, upstream index verification, signed driver validation | Local package integrity checks available; fork signing pending |
| PS.3 | Archive source revision, locks, build logs, inventories, artifacts and rollback | Local evidence retained; durable release archive and standard SBOM pending |
| PW.1 | Threat model for LocalSystem service, API, update and routing boundaries | Initial model documented; independent design review pending |
| PW.2 | Security-sensitive design review before a stable release | Independent reviewer not assigned |
| PW.4 | Review dependency provenance, licensing and vulnerability reports | Lockfiles retained; inherited audit findings require triage |
| PW.5 | Loopback-only Tor endpoints, explicit errors, disabled hosted APIs, no direct proxy fallback | Targeted regression tests available; broader code review pending |
| PW.6 | Record compiler/tool versions and build flags; hardened reproducible release build | Development builds checked; release hardening and reproducibility review pending |
| PW.7 | Go vet, source review and regression checks | Targeted analysis available; full SAST/secrets scan pending |
| PW.8 | Targeted tests plus real Tor/WireGuard, IPv4/IPv6, DNS, driver and update tests | Unit/build evidence available; tunnel and driver integration matrix pending |
| PW.9 | Binary updates and hosted SPN disabled; installer loopback API; settings documented | Initial defaults set; upgrade/default compatibility tests pending |
| RV.1 | Review advisories and dependency scans at changes and releases; accept reports | Manual intake only; private reporting deferred |
| RV.2 | Triage severity, exploitability, affected releases, mitigation and fix in risk register | Initial register created; dependency findings not resolved |
| RV.3 | Record root cause, related affected paths, regression test and process correction | Template available; incident response execution not yet demonstrated |

## Enforcement

Run `scripts/security-check.ps1` and retain the evidence directory. A nonzero exit means a failed gate. Dependency scans can also exit nonzero due to vulnerabilities; preserve and triage the report. Do not downgrade failure to success by creating a baseline without an owner, rationale, review date, and expiry. CI workflow configuration and branch rules must be installed and verified on the actual Git server before calling enforcement operational.

Public stable release is blocked by unresolved reachable critical/high vulnerabilities, missing independent review of privileged changes, absent authenticated release provenance, and untested claimed routing protections. Document findings even when a development build is usable. Experimental local installation does not waive these stable-release criteria.
