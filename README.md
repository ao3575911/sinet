# SINet

**Your terminal. Your network.** A Go social CLI and PostgreSQL API prototype.
The executable remains `project-name` as specified in the supplied contracts.

## Working prototype

- Guided CLI sign-in and keyboard dashboard; scriptable JSON commands on Linux/macOS/Windows.
- Actual development email OTP through Mailpit, private sessions and atomic unique handles.
- Public profiles/share cards/web link previews, directory and membership-protected lobby.
- Reports/blocks, private persistent notes/editor, settings, paginated export and deletion.
- Reviewed connector catalog, schema-validated publishing queue, pinned installation,
  explicit grants/revocation, safe echo through the personal agent, one-use consent and audit.
- Optional OpenAI-compatible chat with explicit data disclosure, honest no-model mode,
  account quotas, privacy-preserving counters and operator moderation/retention commands.

This is a production-oriented **prototype**, not a public-launch release. Validation has
passed PostgreSQL integration/race tests, the real CLI/SMTP journey and a POSIX PTY journey.
See [evidence and limits](docs/IMPLEMENTATION.md) and [operating instructions](docs/OPERATIONS.md).
A live model and real remote connector have not been provisioned. The vulnerability database
lookup is blocked by this environment's network policy and is not counted as passed.

## Start locally

Requires a current patched Go compiler (validation target **Go 1.27.2**) and Docker Compose.

```sh
docker compose up -d --build
make build
./bin/project-name health
./bin/project-name
```

Read OTPs in Mailpit at http://127.0.0.1:8025. The API is http://127.0.0.1:8080.
For native API development: `docker compose up -d db mailpit`, then `./bin/server`.
All published development ports bind to loopback. The server applies versioned,
checksum-protected migrations automatically. Environment variables configure the native
server; it does not automatically source `.env`. See `.env.example`.
Proxy-intercepted cloud builds need the documented [CA mount](docs/OPERATIONS.md).

```sh
./bin/project-name login request you@example.test
./bin/project-name login verify you@example.test  # reads the code from stdin
./bin/project-name claim your_handle
./bin/project-name room lobby join
./bin/project-name plug info echo 1.0.0
./bin/project-name plug install echo 1.0.0
./bin/project-name plug grant echo 1.0.0 repeat --approve
./bin/project-name agent call echo 1.0.0 repeat '{"text":"Hello SINet"}'
./bin/project-name plug revoke echo repeat
./bin/project-name share
```

`project-name help` lists every command. Separate sessions use `PROJECT_NAME_CONFIG`.
Echo is a deterministic tool. No model configuration means “AI unavailable”, not invented
AI output. No shell, USB access, uploaded-code execution, live inbox or wallet is shipped.

## Validate

```sh
make check
make integration       # PostgreSQL on localhost:5432; TEST_DATABASE_URL overrides
make smoke             # running API + PostgreSQL + Mailpit required
make smoke-tui         # POSIX PTY guided onboarding and note editor
make cross             # Linux amd64, macOS arm64, Windows amd64 CLI binaries
```

Integration tests create/drop isolated schemas. `make test` without TEST_DATABASE_URL
explicitly skips PostgreSQL tests and does not prove acceptance. The [CI workflow](.github/workflows/ci.yml)
runs real database/race tests, cross-builds, Mailpit/CLI/TUI journeys and govulncheck.
Live production SMTP/model tests, sustained-load testing and an independent security review
remain release gates. Connector ownership review is an explicit operator attestation;
no real remote service is approved by this repository.

## Contracts and provenance

Active contracts: [PRD](docs/PRD.md), [architecture](docs/ARCHITECTURE.md),
[security](docs/SECURITY.md), [roadmap](docs/ROADMAP.md), [decisions](docs/ADR/0002-implementation-boundaries.md).
Original materials remain unchanged under [reference/](reference/README.md).

The configured remote is `https://github.com/ao3575911/sinet.git`. Milestone commits require
SSH signatures. The owner registered the dedicated public signing key. Local signature
verification passes; remote verification is recorded in docs/IMPLEMENTATION.md when run.
