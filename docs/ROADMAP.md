# Implementation roadmap

## Active follow-up: public GitHub connector

The [next-slice plan](NEXT_SLICE.md) targets a read-only public repository outcome
for 0.2.0. The [standalone gateway](GITHUB_CONNECTOR.md) implements the initial
fixed-repository contract for local review. Live ownership/TLS approval and the
two-account grant/consent/audit/revocation journey remain delivery gates.

## M0: runnable skeleton

- `go.mod`, CLI and server entry points, config loading, health endpoint, Docker Compose with Postgres + Mailpit, migrations and make targets.
- Exit: clean checkout boots local stack and CLI can reach API.

## M1: identity + activation

- OTP, account/session, atomic handles, profile read/edit, minimal shareable web profile/link preview, claim rules, onboarding TUI.
- Exit: two independent users can sign in; collision and session tests pass.

## M2: social + workspace

- Public directory, one moderated room, reports and blocks, per-user notes, settings and shareable text identity cards.
- Exit: messages and notes are scoped; blocked users and invalid access rejected.

## M3: connector hero loop

- JSON Schema contract, curated registry, publisher manifest submission/review, install/uninstall, per-operation grants, `echo` example, audited vetted remote HTTPS adapter (with SSRF tests), audit trail, minimal private agent conversation + deterministic tool invocation.
- Exit: fresh user runs one end-to-end connector interaction; attempted unauthorized tool call is denied.

## M4: actual hosted model + hardening

- Configurable provider API, quotas, usage logs without sensitive prompts, rate limits, backup strategy, packaging, CI, documentation, privacy policy draft.
- Exit: working hosted model demo when configured; honest unsupported state otherwise; security matrix passes.

## Later tracks (not prerequisites)

- Business namespaces and verified-domain mail, local terminal/USB bridge, cloud sandbox, remote MCP adapters, wallet + ENS/DID integration, mobile/web companion, payments.

**Release rule:** no progression by screenshots alone. Require command transcript, test output, migration state and identified known limitations.
