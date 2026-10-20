# ADR 0001: Lean desktop terminal, governed connectors

**Status:** proposed for initial implementation

## Decision

Use Go for CLI and API, Postgres for user-visible state, Mailpit for development email, and a centrally governed declarative connector catalog. Cloud AI uses an explicitly configured model provider. The MVP does not execute untrusted uploaded code or offer an unrestricted hosted shell.

## Why

Single language, portable compiled binary, fewer moving parts, transactional name claims, simpler audits. A curated connector pipeline makes the promised central feature demonstrate real value without pretending arbitrary code uploads are safe.

## Alternatives rejected now

- Full desktop Electron wrapper: shipping weight and complexity before validating terminal retention.
- Smart-contract handle registry: high friction and undesirable speculative claims for ordinary users.
- Full VM per user by default: high infrastructure cost; not needed to validate marketplace appeal.
- Direct `stdio` execution of marketplace-submitted MCP servers: unsafe arbitrary code on host.
- Custodial wallet + full inbox at signup: introduces cost and regulatory/security scope unrelated to hero loop.

## Revisit when

Validated retention and connector adoption justify extra resources, and isolation has measurable proof.
