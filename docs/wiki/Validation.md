# Validation evidence

The Go 1.27.2 suite passes 17 tests with the race detector: nine real PostgreSQL integration
tests and eight unit/adapter checks. Real two-account CLI/Mailpit and POSIX PTY journeys pass.
Linux amd64, macOS arm64 and Windows amd64 builds pass; native non-Linux runtime UX is untested.

CI's verbose govulncheck v1.9.0 reports no vulnerabilities after updating dependencies.
The Windows x/sys advisory GO-2026-5024 was addressed by v0.49.0. Formatting, vet, module
checksums, migration checksums and a real backup restoration pass.

Read the [current evidence ledger](https://github.com/ao3575911/sinet/blob/main/docs/VALIDATION.md)
and [Actions runs](https://github.com/ao3575911/sinet/actions). Tests prove the documented
prototype boundaries; they do not replace independent review or production load testing.

[Documentation home](Home.md)
