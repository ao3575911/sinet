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
| Vulnerability lookup | NOT completed locally: both scanner versions hit a network-policy Forbidden for vuln.go.dev; CI runs v1.9.0 separately |
| Live infrastructure | Production SMTP, model credentials, controlled alias domain/routing and approved real remote service absent |

The clean-container bootstrap and current GitHub CI results are appended below when run.

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

1. Complete vulnerability scanning from a runner allowed to access the official database;
   fix findings before a public release, and independently review trust boundaries.
2. Configure/test production SMTP, TLS, database credentials and a real model with provider
   billing limits. A controlled mail domain/routing is needed before any alias claim.
3. Verify hostname ownership and review one real remote connector; implement upstream
   idempotency/reconciliation before allowing valuable irreversible remote operations.
4. Run sustained-load/spam tests, tune edge limits, implement an account suspension workflow,
   and verify backup encryption, expiry and deletion handling.
5. Validate macOS/Windows interaction and Windows credential ACLs, add OS keychain storage,
   then finalize naming, packaging and release signing/key lifecycle.
