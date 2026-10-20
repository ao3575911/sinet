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
