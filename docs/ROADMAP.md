# Implementation roadmap

## 0.2.0 release candidate: wallet, email and public GitHub

At the owner's request, implement noncustodial wallet ownership proofs and private
email drafts with a user-reviewed Gmail handoff, before the other sections. See
[wallet/email guide](WALLET_EMAIL.md). The source also exposes accurate setup
states for business, ENS/DID identity, devices/storage, terminal/cloud and mobile.
Google OAuth mailbox sync, domain mail provisioning and live wallet-extension
validation remain separate work; no replacement 0.1.1 registry package is published.

The public GitHub integration is complete locally in **0.2.0-rc.1** source. The CLI/TUI
reuses reviewed registry installation, exact-operation grants, expiring single-use
consent, deterministic personal-agent invocation and audit/revocation. Counted fixtures
prove both accounts and default deny; they are not live GitHub evidence. Gateway output
is revalidated server-side. See [guide](GITHUB_CONNECTOR.md), [sprint](GITHUB_SPRINT.md)
and [release-candidate evidence](releases/0.2.0-rc.1-validation.md).

Live delivery remains blocked on controlled public HTTPS hosting, ownership evidence
and operator approval. [Deployment preparation](GITHUB_DEPLOYMENT.md) is ready; no
placeholder endpoint is approved. No release/tag/package publication occurred.

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

The public GitHub slice advances issue #4 locally. Its live two-account journey remains
an external gate. npm/PyPI 0.1.1 publication is verified; those immutable packages lack
the new source features. Marketplace remains part of issue #14.

- Business namespaces and verified-domain mail, local terminal/USB bridge, cloud sandbox, remote MCP adapters, wallet + ENS/DID integration, mobile/web companion, payments.

**Release rule:** no progression by screenshots alone. Require command transcript, test output, migration state and identified known limitations.
