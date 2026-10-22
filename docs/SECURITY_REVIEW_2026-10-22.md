# Security review — 2026-10-22

Reviewed the recent connector changes (`d017f7b..aefa650`) first, then the active
first-party codebase at merged main. Implemented hardening on
`codex/security-hardening`. No model credentials, real email delivery, live wallet
transactions or live connector approval were used. This review does not constitute
an independent penetration test or production certification.

## Findings and resolution

Severity describes this prototype's actual exposure, not a CVSS score.

| Finding | Severity / scope | Evidence and fix |
|---|---|---|
| Consent fingerprints rounded JSON numbers through float64 | High for generic numeric connector inputs; fixed GitHub strings unaffected | On baseline, consent for 9007199254740992 authorized 9007199254740993. Preserve number literals in validation, fingerprinting, CLI requests and responses; changed-number regression now returns 403 with zero executions. |
| Session expired while waiting on account lock still authorized writes | Medium; valid session required at request start | Baseline queued request returned 200 after expiry. Recheck with `clock_timestamp()` before idempotency/handler; regression now returns 401 with zero notes written. |
| Concurrent publishers could claim different versions of one connector ID | Medium; two activated accounts required | Baseline synchronized submissions both returned 200. Connector-ID advisory locking serializes ownership check and insert; now exactly one 200 and one 409. Operator approval remains separate. |
| Production database TLS guard checked a substring | Medium; configuration error or malicious configuration required | A password or unrelated query value could contain `sslmode=verify-full` while effective mode was `disable`. Parse PostgreSQL URLs; require exactly one actual verify-full setting; reject duplicate/malformed parameters and query-host overrides (lib/pq disables TLS on Unix sockets). Public HTTPS origin also rejects embedded credentials/path/query/fragment. |
| Duplicate JSON keys accepted across API/manifest/generic arguments | Medium defense in depth against parser disagreement | Baseline duplicate note fields returned 200. Strict decoding rejects duplicate keys at every object level, trailing values and invalid UTF-8; shared depth/number bounds prevent excessive nesting/numeric resource use. This is exact-key checking; clients should use the documented lowercase field names. |
| Upstream body timeout/truncation misreported as size limit | Low; recent diff failure-path gap | Isolated readers produce timeout or unexpected EOF. Gateway now returns fixed 504 timeout / 502 failure; adapter returns fixed timeout / invalid-gateway codes. No remote body is exposed. |
| Database errors looked like expired credentials or exhausted rate limits | Low; operational reliability | Forced missing-table fixtures returned misleading 401/429 on baseline. Auth/rate storage failures now return 503; failed rate commits cannot claim throttling. Public-profile DB errors return 503 rather than false 404. |
| Untrusted error text and malformed local credentials insufficiently bounded | Low; hostile API or local config required | Error response budget is now 4096 bytes and only `[a-z][a-z0-9_]{0,127}` is displayed. Existing terminal control filtering was already present. Session files must be regular, <=4096 bytes, retain Unix permissions/origin checks, and contain a 64-character hex token. Failed atomic save no longer changes in-memory authentication. |

Additional preventive hardening: parameter schemas are limited to inline trees
(8192 bytes, 512 value nodes, depth 16); all reference keys are denied, including
local/dynamic/recursive references. This prevents small reference graphs from
amplifying validation work and does not rely on request-context cancellation to
interrupt CPU work. General JSON nesting is capped at 64, number literals at 128
bytes, explicit exponents at -308..308. These are conservative application limits,
not a claim that every previously accepted schema was exploitable. Health/profile
DB requests have deadlines; SMTP closes its connection on caller cancellation and
honors an earlier deadline. Wallet browser recovery handles a provider rejecting
with `null`, retaining a usable retry button and bounded message.

## Review coverage

- API identity/OTP/session transactions; route/body/idempotency handling; all social,
  notes, profile, reporting, export/delete and account-filtered queries; wallet/email
  challenges, draft access and deletion; registry submission/review/install/grants,
  consent, deterministic and model calls, audit and quotas.
- Gateway fixed inputs and fields, response validation, request budgets, DNS and
  connected-peer checking, redirects and TLS; CLI disclosure, rendering, input paths,
  credential storage and error handling; model and SMTP adapters.
- Wallet signature recovery, canonical signatures, origin/nonce/expiry binding and
  browser CSP/DOM handling; Gmail URL encoding and explicit handoff. No automatic
  tool execution or mailbox access was introduced.
- All migrations and the migration runner, operator commands, embedded contracts,
  Docker/Compose resources, active and historical npm/PyPI launchers, packaging and
  publication scripts, repository Actions and signing/attestation trust boundaries.
  Reference originals/extracted materials and immutable historical migrations/tags
  were preserved. No SQL injection or cross-account data access was identified in
  the reviewed parameterized query paths; existing negative tests remain necessary.
- Workflows keep PR execution separate from metadata-only `pull_request_target`
  labeling; release workflows require signatures/provenance and approved source.
  Existing version-tagged actions/images remain mutable upstream dependencies;
  pinning and regularly updating reviewed digests is a remaining supply-chain task.

## Reproduction and validation

An isolated detached baseline checkout at `91b7056` received only the new API
regression tests. Before hardening, the first four tests below failed with:

```
TestSecuritySessionExpiryAfterLockWait: expired queued session accepted: HTTP 200
TestSecurityConsentExactNumbers: expected HTTP 403 got 200
TestSecurityConcurrentConnectorOwnership: multiple publishers claimed ID: 200 200
TestSecurityAmbiguousRequestsDenied: expected HTTP 400 got 200
```

The fixed tests exercise real PostgreSQL transactions, observe lock waiters, count
executions and check persisted state. They do not contact GitHub. The numeric
connector uses synthetic review evidence only inside a disposable test schema.

Checks observed locally on 2026-10-22 using Go 1.27.3, isolated PostgreSQL 18.6,
Mailpit 1.27.10 and loopback-only development listeners:

- `make check integration cross`: fmt, vet, module verification, all race tests and
  Linux amd64/macOS arm64/Windows amd64 builds passed. Tests include both-account
  GitHub consent/expiry/replay/revoke/isolation, counted transport, real CLI/PTY,
  malformed/private/oversize responses and existing DNS/peer-IP/redirect defenses.
- New fault tests cover expired queued sessions, exact numeric consent, publisher
  concurrency, duplicate JSON, production URL validation, schema limits, bounded
  client errors/credentials, upstream body failures and SMTP cancellation.
- Forced auth/rate-table failures return truthful 503 responses.
- CLI/SMTP two-account smoke, POSIX PTY smoke and Chromium wallet regression passed.
  Browser fixtures reject wrong address/chain/message and handle null provider
  errors, with no external requests. Gmail handoff creates a URL only.
- 15 Python packaging/version/publication fixture tests, 2 Node automation tests,
  actionlint 1.7.12 with ShellCheck, and compiled CLI/version agreement passed.
- Strict JSON fuzzing: 15-second run, 22,688 executions, no failure. This bounded run
  is not an exhaustive parser proof.
- `govulncheck@v1.9.0 -show verbose ./...`: zero reachable findings and zero findings
  in imported packages. GO-2026-5932 remains at module inventory level for unused
  `golang.org/x/crypto/openpgp`; wallet code uses `sha3`, not OpenPGP.

Reproduce application checks with a fresh isolated database and Mailpit:

```sh
GOTOOLCHAIN=go1.27.3 make check integration cross \
  TEST_DATABASE_URL='postgres://sinet@127.0.0.1:15432/sinet?sslmode=disable'
GOTOOLCHAIN=go1.27.3 make build
# Start ./bin/server with DATABASE_URL above, APP_HTTP_ADDR=127.0.0.1:18080,
# APP_PUBLIC_URL=http://127.0.0.1:18080 and SMTP_ADDR=127.0.0.1:11025.
SINET_API=http://127.0.0.1:18080 MAILPIT_URL=http://127.0.0.1:18025 python3 tests/smoke.py
SINET_API=http://127.0.0.1:18080 MAILPIT_URL=http://127.0.0.1:18025 python3 tests/tui_smoke.py
SINET_API=http://127.0.0.1:18080 MAILPIT_URL=http://127.0.0.1:18025 python3 tests/wallet_browser.py
```

Five unchanged migrations (0001–0005) were applied and checksums matched the prior
sprint. After both smoke runs, the isolated database contained zero accounts and zero
test schemas. Only review-owned services/data are stopped and removed during cleanup;
the original root-owned working tree is untouched. The editable review checkout is
`/tmp/sinet-connector-sprint`.

## Compatibility and remaining gates

Existing bundled manifests meet the new schema limits. Other manifests using
references must be inlined and submitted under a new immutable version; old versions
fail closed. Pending numeric consents may require renewed approval because their
fingerprints now preserve exact literals. No database migration or version bump is
needed for this unreleased 0.2.0-rc.1 hardening; published 0.1.1 packages stay unchanged.

The local Docker daemon is unavailable to this Unix user. Current-branch container
validation must be established in CI; the prior sprint's result is not new evidence.
Native macOS/Windows runtime and Windows ACL behavior were not exercised locally.
Trusted private config parent directories are still required; local account compromise
and arbitrary concurrent replacement of parent directories are outside the file
checks. OS keychain integration remains pending.

No controlled hostname, hosting access, ownership evidence or operator approval
exists for the public gateway. The live GitHub journey remains **unverified**.
Provision controlled HTTPS hosting and complete the separate documented operator
review, then record a real two-account journey. Public launch also still needs the
independent review (#2), production mail/TLS/model setup, abuse/edge rate controls,
load/suspension/backup recovery evidence and native-platform checks (#3–#6).
There is no distributed transaction with a remote service: failure after outbound
execution but before database commit can leave its outcome uncertain. Do not approve
remote writes on the strength of these read-only GitHub fixtures.

## Requested merge-history normalization

The owner requested that the merge subject match the existing conventional commit
style. Replaced only merge `6da4b8c` with signed `91b7056bc3d6af28f199fe844e9a9f601f583f21`:
`feat: complete governed GitHub connector and prepare 0.2.0 RC (#19)`.
The tree, original author and both parents are unchanged; an exact-SHA force-with-lease
protected main from concurrent updates. Local signature verification and GitHub
`verified: true, reason: valid` passed. Local recovery reference:
`recovery/main-before-merge-message`. No historical tag was rewritten.
