# SINet

**Your terminal. Your network.** A Go social CLI and PostgreSQL API prototype.
The project and executable are named **SINet** (case sensitive).

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
A live model and real remote connector have not been provisioned.
[GitHub CI](https://github.com/ao3575911/sinet/actions/runs/39351660041) passes the full
journey and govulncheck v1.9.0 reports **no vulnerabilities found** on the updated dependencies.

## Start locally

Requires a current patched Go compiler (validation target **Go 1.27.2**) and Docker Compose.

```sh
docker compose up -d --build
make build
./bin/SINet health
./bin/SINet
```

Read OTPs in Mailpit at http://127.0.0.1:8025. The API is http://127.0.0.1:8080.
For native API development: `docker compose up -d db mailpit`, then `./bin/server`.
All published development ports bind to loopback. The server applies versioned,
checksum-protected migrations automatically. Environment variables configure the native
server; it does not automatically source `.env`. See `.env.example`.
Proxy-intercepted cloud builds need the documented [CA mount](docs/OPERATIONS.md).

```sh
./bin/SINet login request you@example.test
./bin/SINet login verify you@example.test  # reads the code from stdin
./bin/SINet claim your_handle
./bin/SINet room lobby join
./bin/SINet plug info echo 1.0.0
./bin/SINet plug install echo 1.0.0
./bin/SINet plug grant echo 1.0.0 repeat --approve
./bin/SINet agent call echo 1.0.0 repeat '{"text":"Hello SINet"}'
./bin/SINet plug revoke echo repeat
./bin/SINet share
```

`SINet help` lists every command. Separate sessions use `SINET_CONFIG`.
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

## Guides and community

- [Published wiki](https://github.com/ao3575911/sinet/wiki) and [versioned guides](docs/wiki/Home.md)
- [Documentation site](https://ao3575911.github.io/sinet/) — built and deployed by GitHub Actions
- [Labelled release work](https://github.com/ao3575911/sinet/issues) and [Discussions](https://github.com/ao3575911/sinet/discussions)
- [Contributing](CONTRIBUTING.md), [security reporting](SECURITY.md), [community expectations](CODE_OF_CONDUCT.md)
- [Repository labels, workflows and signed publishing](docs/REPOSITORY.md)
