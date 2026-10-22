# Current candidate evidence

The **0.2.0-rc.1** [22 October 2026 ledger](releases/0.2.0-rc.1-validation.md) supersedes
current-capability claims below: Go 1.27.3, five migrations, counted two-account GitHub
CLI/PTY fixtures, real SMTP smoke and wallet browser regression pass locally. Live
GitHub hosting/approval remains blocked. The patched scan has zero reachable findings
and one unused-package module advisory. Earlier records below retain their original
scope, compiler/date and registry publication state; they are historical evidence.

# Prototype validation — 20 October 2026

This is a command/test evidence ledger, not a production certification.

| Check | Result and scope |
|---|---|
| Source inventory | No application code initially; contracts copied to active root paths; preserved source library unchanged |
| Source provenance | All 107 reference/MANIFEST.json SHA256 checksums verified |
| Compiler | Go 1.27.2, official download SHA256 verified against go.dev metadata |
| Formatter / static checks | `make check`: fmt, vet, module checksum verification passed |
| Unit and integration / race | `TEST_DATABASE_URL=... go test -race ./... -count=1`: 17 passed, 0 failed, 0 skipped |
| Real PostgreSQL | Nine integration tests use independent schemas, real transactions/constraints and forward migrations |
| Identity abuse | OTP expiry/replay/attempt limit, request throttling, concurrent redemption, session expiry/logout, reserved/homograph handles and simultaneous claim checks pass |
| Isolation / data | Private profiles, note reads/edits/deletes, room membership, block/report behavior, chat/audit separation, settings, export pages and deletion pass |
| Connector authorization | Review state, immutable versions, grants, input validation, tenant/version checks, revocation, quotas and audited echo loop pass |
| Consent / prompt boundaries | Input/account/version binding and one-use consent pass; chat cannot create grants or auto-execute tools |
| Egress | Private/reserved/transition IP denial, mixed DNS answers, pinned dial, connected-peer mismatch and redirects tested |
| Model adapter | Configured HTTP fixture, output budget, upstream error redaction, redirect denial and explicit no-model/disclosure behavior tested; no live AI demo |
| Operator / persistence | Approval attestations, report resolution/removal, maintenance, repeated migrations and checksum-tampering denial pass |
| Native CLI + SMTP | Two-account journey passed 51 CLI checks with real Mailpit delivery; no OTP/token printed |
| POSIX TUI | PTY signup, hidden OTP, handle, dashboard, note save, arrows and exit passed after correcting coalesced-key handling |
| Portable binaries | Go 1.27.2 CLI builds passed for Linux amd64, macOS arm64, Windows amd64; Windows/macOS runtime not tested |
| Backup restore | PostgreSQL custom dump restored into separate sinet_restore_validation; 3 migration records, 1 connector, 1 persisted note matched source counts |
| Vulnerability scan | CI run 39351660041: govulncheck v1.9.0 with verbose output reports no vulnerabilities found across all 5 modules and Go 1.27.2; local lookup is policy-denied |
| Live infrastructure | Production SMTP, model credentials, controlled alias domain/routing and approved real remote service absent |

Clean-container bootstrap passed before the naming change; final renamed image results follow below.

## Useful commands

```sh
make check
make integration
make smoke
make smoke-tui
make cross
```

`make integration` supplies the local PostgreSQL URL; `make test` without
TEST_DATABASE_URL skips database tests and is not sufficient evidence. Smoke scripts need
running API/Postgres/Mailpit. Override SINET_API and MAILPIT_URL for isolated stacks.
See OPERATIONS.md for full commands, ownership-review requirements and resource limits.

## Immediate release work

1. Obtain an independent security review of identity, consent, tenancy and egress; keep
   dependency scanning in CI and review new advisories before release.
2. Configure/test production SMTP, TLS, database credentials and a real model with provider
   billing limits. A controlled mail domain/routing is needed before any alias claim.
3. Verify hostname ownership and review one real remote connector; implement upstream
   idempotency/reconciliation before allowing valuable irreversible remote operations.
4. Run sustained-load/spam tests, tune edge limits, implement an account suspension workflow,
   and verify backup encryption, expiry and deletion handling.
5. Validate macOS/Windows interaction and Windows credential ACLs, add OS keychain storage,
   then finalize packaging and release signing/key lifecycle.

## Final dependency and naming evidence

- The executable is **SINet**; CLI environment variables are SINET_API / SINET_CONFIG.
- GitHub CI [39351660041](https://github.com/ao3575911/sinet/actions/runs/39351660041)
  passed at source commit `4604df255e549c12d2c56bef1a54301ba0aaa2d6`: formatting, vet,
  module integrity, race/database suite, all three builds, actual CLI/SMTP and PTY journeys,
  and the verbose vulnerability scan. The earlier Windows advisory GO-2026-5024 was fixed
  by upgrading x/sys to v0.49.0 (fixed upstream from v0.44.0).
- Scanner module inventory: lib/pq v1.12.3, jsonschema/v5 v5.3.1, x/sys v0.49.0, x/term
  v0.47.0 and the local module; Go 1.27.2 standard library. Result: **No vulnerabilities found.**
- Fresh-image bootstrap previously verified nonroot UID 65532, read-only root filesystem,
  all capabilities dropped and no mounts. The updated image retains those controls.
- Source originals and historical migrations remain unchanged; 0004_sinet_brand.sql updates
  existing databases. Only historical reference/migration content retains the generic name.

## Startup race regression and wiki

A subsequent empty-volume test found an initialization-readiness race; it was fixed,
not counted as a pass. The corrected TCP health check and bounded API retry then passed
against a newly created sinet-ready volume. Real CLI/SMTP and PTY tests passed; database
checks confirmed 4 migrations, publisher SINet and zero accounts after cleanup. See #7.

The wiki guides were actually published at https://github.com/ao3575911/sinet/wiki with
signed commit e6ea1759fe54c0a3de241cb036a117492f210cb8 and exact remote-SHA verification.
Two repository-automation tests also pass; actionlint v1.7.12 validates the workflows.

## Repository checks and main delivery

[Validation run 39352937595](https://github.com/ao3575911/sinet/actions/runs/39352937595)
passed at commit 1a5926fc13d8d712818e96091b6ff74788a28954: both automation tests, actionlint
including ShellCheck, 17 race-enabled Go tests, real SMTP/CLI/PTY journeys, portable builds
and vulnerability scanning. The previously failing unused-variable ShellCheck warning is
fixed. Pages and community publishing are activated by the authorized main push; their
actual results are recorded after the workflows complete.

## Published repository features

At main commit 466447fd354d59b465bba01802a13c457aad5fd1, all five workflows passed:
validate (39353107801), dependency-security (39353107712), portable-binaries (39353107323),
community-setup (39353107429) and documentation-pages (39353107401). GitHub marks all eleven
implementation commits through that point verified=true, reason=valid. The remote main
SHA matched local HEAD; PR #1 is merged.

Issues/PR custom labels were read back, and Discussions #9, #10 and #11 plus their labels
were verified in published GitHub HTML. Wiki guide HTML was also verified. Pages deployed
to https://ao3575911.github.io/sinet/; the restricted runtime receives 403 on github.io,
so an explicit post-deployment HTML/navigation check was added to the Pages runner.
Portable artifact 12234416437 includes three builds/SHA256SUMS with 14-day retention.

SemVer additions validate VERSION=0.1.0, a dated changelog and CLI agreement, with three
Python tests. Release-tag signature checks run against the public main-branch trust list.
The actual signed tag result is recorded after publication; preparing a tag is not reported
as a verified release until those checks pass.

## Portable release packaging

The 0.1.0 packaging rehearsal built three CLI archives and the Linux server archive with
CGO disabled. SHA256SUMS validated all four archives and BUILDINFO.json. Archive inspection
found no absolute or traversal paths; the extracted Linux CLI reported version 0.1.0.
Three SemVer tests and actionlint passed. macOS/Windows remain cross-build-only evidence.
The signed-prototype-release workflow adds signed-tag/commit checks, integration tests,
OIDC build provenance verification and remote asset digest checks before public publication.

Tag v0.1.0 was verified by GitHub, but Actions checkout fetched the peeled commit over
its local tag reference. Both signature-check jobs rejected that local lightweight ref
before packaging/publication. Version 0.1.1 explicitly fetches the actual remote tag and
requires HEAD to equal its signed commit. The original tag is retained, never moved.
Binary packages also include Go/runtime dependency license notices.

The v0.1.1 build passed trusted source signatures, PostgreSQL integration/race tests,
vulnerability scanning, packaging and OIDC provenance/checksum verification. Publication
paused at draft verification: GitHub's release-by-tag endpoint returned 404 for the draft.
The publisher now uses the authenticated draft-inclusive listing. Two regression tests
verify successful draft finalization and rejection of mismatched digests; all five Python
tests and actionlint pass. A signed publication intent requests recovery of those exact
already-attested assets, with no asset replacement.


## Published prototype verification

GitHub prerelease **v0.1.1** is public, not draft, with six assets. Its annotated tag object
is 9b7dd66712492e36085d480df50de1baff827144, pointing at signed release commit
0ef98b111f445bd6ed95874ba630577d0a0d500f; GitHub reports the tag signature verified/valid.
Finalization run [39364975856](https://github.com/ao3575911/sinet/actions/runs/39364975856)
passed intent/tag/commit verification, exact-source provenance, checksums and all remote
asset digests before publishing. The release source's full CLI/SMTP/PTY validation passed
in run [39364592136](https://github.com/ao3575911/sinet/actions/runs/39364592136).

An independent public HTTPS download of SINet-0.1.1-linux-amd64.tar.gz matched SHA-256
`a3d1de8c480abe462389df94fbb51f2e750e3702ea854d492a287301fe62cdbb`.
The extracted binary reported 0.1.1, and the archive included dependency license notices.
macOS/Windows have cross-build evidence only. No npm/PyPI/Marketplace publication is claimed.
Those channels need owner-controlled registry projects/authentication and tested installers
or an appropriate Action listing; see issue #14. No registry credentials were supplied.

## Registry distribution preparation

Prepared @ao3575911/sinet and sinet-cli 0.1.1 candidates from public release bytes, with
the release manifest checked against authenticated GitHub metadata. Linux npm global
installation, npx invocation, pip wheel installation, version checks and both launchers'
tamper denials passed. All three platform wheels passed Twine 6.2.0 strict metadata
validation; license metadata matches the repository's Apache-2.0 license. Launchers add
no runtime libraries or download hooks. [Registry CI 39366222801](https://github.com/ao3575911/sinet/actions/runs/39366222801)
passed native Linux and Windows npm/npx/pip installs and both tamper denials; macOS remains
cross-build-only evidence. These are candidate tests, not registry publication.


## PyPI publication and npm access gate

[PyPI publication run 39368811306](https://github.com/ao3575911/sinet/actions/runs/39368811306)
completed successfully on its second attempt. The first attempt stopped during Sigstore
Rekor signing with HTTP 502, before upload; signing remained mandatory on retry. All three
0.1.1 wheels are public through Trusted Publishing, with matching candidate hashes:

| Wheel platform | SHA-256 |
| --- | --- |
| macosx_14_0_arm64 | becdff2f1bf04a6b4b3497150980a86788de14ee5f8e33e05c54939561773c65 |
| manylinux_2_28_x86_64 | 375c42349f9479ce51d5eb7e545d39dfe0774ff6da8059fc7d5016a4da1e5ee7 |
| win_amd64 | 7b72d68b7ddd04b5c7baf5cf79b40f53aa4f7f18d0937c8a3ca88c34d939bc80 |

An independent public PyPI install in this workspace matched all three hashes to the
CI candidate artifact and the installed Linux command reported 0.1.1. Candidate artifact
12240836406 was downloaded and its ZIP SHA-256 matched authenticated GitHub metadata:
124f6c9b7929116dd5814770cc67d31b02ccb7ed4722e03601abc370bf365400.

The first npm intent stopped before upload because npm interpreted an archive path without
`./` as a GitHub shorthand. The corrected explicit local path passed a publication dry-run;
a signed retry intent retained the original immutable tag. In
[npm retry run 39369067573](https://github.com/ao3575911/sinet/actions/runs/39369067573),
Linux/Windows installs and provenance verification passed; npm signed provenance but the
registry rejected PUT @ao3575911/sinet with E404 (not found or no permission). No npm
package is claimed. Scope ownership/token access remains an external account-owner gate.
Main remains unchanged until both requested registries are published and verified.

All 13 Python version/publication guard tests and actionlint pass after the path/retry fix.

## Security hardening follow-up — 2026-10-22

See [the security review ledger](SECURITY_REVIEW_2026-10-22.md) for baseline failures,
fixed regression coverage, full application checks and outstanding external gates.
This supersedes older test-count summaries without changing their historical results.
