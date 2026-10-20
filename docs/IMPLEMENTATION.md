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
is byte-limited. Public profile responses and test transcripts exclude session tokens and OTPs.

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

Final local Go 1.27.2 suite: **17 tests passed, 0 failed, 0 skipped**, with `-race`:
nine PostgreSQL integration tests and eight unit/adapter security tests. Three portable
CLI builds passed on Go 1.27.2. Initial GitHub CI failed during container initialization
because the runner parsed a single-quoted health command incorrectly; health-command
quoting was corrected to the runner's supported double-quoted form. No application test
had run in that failed workflow. The PTY harness now honors MAILPIT_URL for isolated stacks.

Owner naming decision: the project and executable are now **SINet**, preserving that case.
Active docs, help, profile/share commands, build targets, examples, runtime policy and
CLI tests were renamed; config variables are SINET_API / SINET_CONFIG. Historical SQL
migrations and reference library remain checksum-stable. Migration 0004 updates existing
bundled echo publisher metadata to SINet. Fresh-install PostgreSQL role/database defaults are now sinet. Existing local volumes
retain their identifiers through explicit POSTGRES_USER / POSTGRES_DB overrides.

Before the rename, final isolated Go 1.27.2 API image bootstrap passed from a new empty
PostgreSQL volume: native CLI and PTY journeys passed on ports 8081/8026. Container
inspection confirmed UID 65532, read-only root, all capabilities dropped and no mounts.
The renamed binary and migration are being validated against that same persistent stack.

GitHub Actions run 39351071802 passed the full Go/race/cross-build, real SMTP/CLI/PTY
journeys and govulncheck v1.9.0. The scan reported zero called-code vulnerabilities,
zero vulnerabilities in imported packages and one advisory in a required module outside
the used code path; dependencies are being updated before main delivery. Local database
lookup remains denied by policy; the allowed CI runner supplied this external evidence.
The SINet rename passed local race/integration checks, all three cross-builds, the real
CLI journey and PTY journey against the existing isolated container stack.

Dependency hardening: updated lib/pq to v1.12.3, x/term to v0.47.0 and x/sys to v0.49.0
without adding dependencies. Minimum module Go version is now 1.26; validation/compiler
remain Go 1.27.2. Formatter/vet/checksums and the full real-PostgreSQL race suite pass with
these versions. The Docker build now uses one allowlisted source-copy layer to reduce
builder storage amplification; .dockerignore continues to exclude credentials and work/.

Final dependency CI scan (39351660041) passed and explicitly reported “No vulnerabilities
found” after scanning the five required modules and Go 1.27.2 standard library. Earlier
module advisory GO-2026-5024 affected x/sys/windows before v0.44.0; v0.49.0 removes it.
The owner's subsequent review explicitly authorizes pushing the completed changes to main.

The final empty-volume bootstrap exposed a temporary PostgreSQL Unix-socket readiness
race. Fixed Compose/CI health checks to require TCP readiness and added a 30-second API
startup retry with a default 5-second connection timeout. Revalidated 17 Go tests with
race detection and a newly created sinet-ready volume: health, the real CLI/Mailpit journey
and the PTY journey passed; 4 migrations applied, publisher SINet, 0 test accounts remain.
Bug #7 records the failure, fix and evidence and is closed as completed.

Repository polish requested by the owner: labelled PR #1 and release-gate issues #2–#6;
issue #8 tracks feature publication. Owner enabled Discussions/Pages and initialized wiki.
Nine wiki guides plus navigation were published in signed commit
`e6ea1759fe54c0a3de241cb036a117492f210cb8`; local signature and exact remote SHA verified.

Repository workflow lint now also runs on GitHub with ShellCheck. Its first remote run
caught an unused retry-loop variable that the local environment's absent ShellCheck did
not inspect; changed it to the conventional ignored variable. Automation tests, actionlint
and Python compilation pass locally. New workflows use narrow permissions and run community
writes only from trusted main/manual configuration; PR labelling never executes PR source.

Workflow check fix verified remotely: run 39352937595 is green for automation tests,
workflow/ShellCheck lint, the full Go validation, cross-build, real Mailpit/PTY journeys
and vulnerability scan. Owner explicitly requested finalization and main delivery.

All five initial main publishing/check workflows passed at 466447f: validate,
dependency-security, portable-binaries, community-setup and documentation-pages. Custom
labels were read back on issues/PRs; discussion starters #9–#11 and their labels rendered
on GitHub. Pages deployment succeeded at https://ao3575911.github.io/sinet/; direct fetching
from this restricted runtime returned 403, so the Pages workflow now checks its own public
HTML from the allowed runner. The wiki is published and signed. Portable artifact
12234416437 contains the three builds and checksums (14-day retention).

Owner requested SemVer with signed verification. Added canonical VERSION=0.1.0, dated
CHANGELOG.md, embedded local CLI version, three Python version tests, and a semantic-version
workflow checking syntax/changelog/CLI and trusted annotated tag/release commit signatures.
The signing trust list contains only the public key; private keys remain outside CI/source.
