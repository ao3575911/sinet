# SINet

**Your terminal. Your network.** A Go social CLI and PostgreSQL API prototype.
The executable remains `project-name` as specified in the supplied contracts.

## Current evidence (20 October 2026)

Implemented and tested against PostgreSQL: email OTP with one-use/expiry/attempt limits,
revocable hashed sessions, atomic normalized handles, public JSON/HTML profiles,
private agent identity, directory, lobby membership/messages, reports and blocks,
private persistent notes, settings, idempotent writes, account export/deletion.

Seven PostgreSQL integration tests pass, including the echo connector hero loop,
review-only submissions, version-pinned installs, grants/revocation, one-use input-bound
consent, private agent history, explicit model disclosure and daily quotas. Manifest and
SSRF/connected-IP/redirect unit tests pass. Tests use capturing mail and model adapters;
actual Mailpit delivery and the full CLI journey are pending. CLI and Docker API
packaging are under construction.
This is not a production release. See [implementation evidence](docs/IMPLEMENTATION.md).

## Development

Requires Go 1.25+ and Docker Compose. The database and Mailpit bind only to loopback.

```sh
docker compose up -d db mailpit
go run ./cmd/server migrate
go run ./cmd/server
```

The server reads environment variables, not `.env` automatically. Defaults match the local
Compose stack. Health: `curl http://127.0.0.1:8080/healthz`. Mailpit: http://127.0.0.1:8025.

```sh
TEST_DATABASE_URL='postgres://projectname:local-development-only@127.0.0.1:5432/projectname?sslmode=disable' go test ./internal/api -v
```

The test suite creates and drops isolated schemas in the test database. Omit
`TEST_DATABASE_URL` to skip database tests explicitly; a skipped suite is not acceptance.

No deployed email domain, hosted model credentials or live remote connector have been
provisioned. No host shell, USB access, uploaded-code execution or inbox is provided.
SMTP failures return an explicit unavailable error. Production setup is a separate gate.

## Contracts and provenance

Active contracts: [PRD](docs/PRD.md), [architecture](docs/ARCHITECTURE.md),
[security](docs/SECURITY.md), [roadmap](docs/ROADMAP.md), [decisions](docs/ADR/0002-implementation-boundaries.md).
Original materials remain unchanged under [reference/](reference/README.md).

The configured remote is `https://github.com/ao3575911/sinet.git`. Milestone commits require
SSH signatures. The owner registered the dedicated public signing key. Local signature
verification passes; remote verification is recorded in docs/IMPLEMENTATION.md when run.
