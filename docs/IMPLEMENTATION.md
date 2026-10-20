# Implementation and evidence

## Plan recorded before implementation

- M0: API health, Postgres/Mailpit Compose, checksummed migrations, CLI health; bootstrap.
- M1: hashed expiring one-use OTP, throttles, sessions, atomic handles, public profiles.
  Verify replay/expiry/brute force, concurrent claims, private fields, logout.
- M2: directory, lobby membership/messages, blocking/reporting, notes/settings/export/delete.
  Verify account and room isolation, pagination, persistence and deletion.
- M3: schema validation, review-only submission, immutable approved versions, installs,
  grants/revocation, safe echo, remote egress, one-use consent, audit and private agent.
  Verify no grants from model text, argument limits, SSRF/redirects and version mismatch.
- M4: optional model, bounded quotas, CLI menu/editor, packaging/CI and operating docs.
  Run fmt/vet/test/race, empty-database migration and full HTTP/CLI journey, cross-build.

Milestone results will be appended as checks run. Source references: docs/TESTS.md and
docs/ROADMAP.md. Architecture choices: docs/ADR/0002-implementation-boundaries.md.

## M0–M2 first validation

2026-10-20: Docker PostgreSQL 17.6 and Mailpit 1.27.10 started. Go 1.25.3 compiler installed
in ignored work/ because the preinstalled `go` executable was the board game, not Go.
`go test ./internal/api -v` with TEST_DATABASE_URL: four tests passed (1.395s).
Checks cover OTP replay/expiry/attempt lockout/request throttling, simultaneous handle
claim, private profile fields, link-preview HTML, room membership, blocks/reports,
cross-account note read/edit/delete denial, note persistence, settings, export/deletion,
handle tombstone, idempotency conflict/replay, invalid pagination and large requests.
No external SMTP delivery, CLI bootstrap, model provider or remote service claimed yet.
Signing blocked: owner chose to provision an existing signer; no signer available yet.

## M3 validation and signing

2026-10-20: PostgreSQL-backed full suite passed: seven API integration tests plus manifest
validation and safe-egress unit tests. Echo hero loop, grants, tenant/version isolation,
revocation, submission review state, consent binding/consumption, model disclosure and
quotas pass. Model test uses a named test provider; no live model demo is claimed.
A dedicated SSH signing key was generated at the owner's request, outside Git in work/.
The owner confirmed registering its public key as a GitHub signing key. Commit eb60b86
verified locally with `git verify-commit` and `%G? = G`.

## M4 intermediate validation

Real native CLI journey against PostgreSQL + Mailpit: 51 allowed/denied command checks
passed, including actual SMTP delivery, OTP replay denial, two accounts, public HTML/link
preview, private notes, shared room/blocks/reports, echo grants/revocation and deletion.
POSIX PTY journey passed guided onboarding, hidden OTP entry, handle claim, keyboard
navigation, multiline note save and exit. A coalesced-key handling issue found by the PTY
test was fixed before completion. Go 1.25.3 race checks and Linux/macOS/Windows builds
passed. Go 1.27.2 was downloaded with SHA256 checked against official release metadata;
fmt/vet/module integrity/race tests also passed on that compiler. The final image and
portable builds now target Go 1.27.2. Additional tests cover operator review attestation,
moderation removal, retention maintenance, migration reapplication/checksum tampering,
concurrent OTP redemption and session expiry. Export was bounded into pages; note input
is byte-limited. No data response or test transcript includes a session token or OTP.

The vulnerability scanner (both v1.1.4 and current v1.9.0) could not query the official
vuln.go.dev database: environment HTTP policy returned Forbidden. This is an uncompleted
check, not a clean result. CI includes a pinned current scanner for an allowed runner.

Signing history correction: GitHub rejected the first unpublished push because the
inherited author email was private. Unpublished implementation commits were reauthored
using the public noreply address from upstream, re-signed and rebased onto current main,
preserving the owner's license/ignore changes. No published history was force-updated.
Final initial milestone IDs: d75c6c6 and 7a779e2. GitHub REST verification reports
verified=true, reason=valid for both (2026-10-20T00:48:45Z). The branch SHA matched
`git ls-remote`. Reference provenance commit d5151fc retains all supplied materials;
all 107 reference/MANIFEST.json checksums passed before committing.
