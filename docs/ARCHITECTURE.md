# Architecture: one executable, explicit trust boundaries

## Recommended stack

- **Client:** Go CLI, Bubble Tea TUI when appropriate, Cobra or standard `flag` for scriptable commands. Avoid GUI framework and mandatory local daemon in P0.
- **Share layer:** small server-rendered public profile page (Go HTML templates and OG metadata), separate from private API.
- **Server:** Go REST/JSON API, secure cookie/device-session or scoped bearer auth, background jobs in the same deployable initially.
- **Database:** PostgreSQL with proper constraints and transactions. Migrate forward via versioned SQL migrations.
- **Development email:** Mailpit SMTP. Production provider adapter with explicit configuration and delivery errors.
- **AI:** provider abstraction for an OpenAI-compatible API, with per-user quotas. If no provider configured, show “AI unavailable” while still permitting deterministic built-in tool actions, without presenting canned output as AI reasoning.
- **Connector registry:** manifest JSON Schema, installed version, explicit grants. Built-in in-process connectors initially. Untrusted executable uploads are out of scope.
- **Packaging:** Docker Compose for API, Postgres, Mailpit; build CLI for Linux/macOS/Windows using reproducible Go builds.

## Request map

```text
Host CLI (TUI + local files only by user command)
   | TLS + authenticated REST
   v
API [email OTP | handles | profiles | rooms | notes | agent | connectors]
   |       |             |              |
Postgres  SMTP adapter  Model adapter   Curated connector runner
                                      (grants enforced by backend)
```

P0 has **no general remote shell or untrusted plugin executor**. A future remote workspace belongs in an independent worker with container/VM isolation, quotas, secrets firewall, network policy, filesystem mounts, and auditable lifecycle. Never rely on the CLI user's terminal theme or prompts as a sandbox.

## Identity model

Account (`id`) owns a normalized canonical handle; the public profile exposes handle, bio, optional avatar and public achievements. Email stays private. Agency is scoped to account; business organizations later receive independent namespace, agent, and billing owner. The account's public handle is not a DID or an ENS name by default.

## Connector model

Manifest: id, version, publisher, description, operations, permissions, privacy metadata, and runtime (`builtin` or `remote_https`). Registry accepts untrusted submitted metadata into review; only audited built-ins or vetted HTTPS services can be enabled. To publish remote services, verify publisher control of the HTTPS hostname, use fixed operation routes, disallow redirects, block localhost/link-local/private/multicast targets at DNS resolution and actual connect time, apply payload/time budgets, and require a trust review. Each call passes through permission checking, explicit consent where needed, input validation, timeout, bounded output, safe egress, and audit record. MCP interoperability is a future adapter, not a claim that all manifests are MCP servers.

## Cloud agent

Agent state is tenant-isolated and stored server side. Model access is optional but explicit. Tools are exposed per installed connector and allowed grants. System instructions must not authorize host-file access, transfers, payments, secrets, or changes to permissions. Tool outputs are treated as untrusted data. The user can list/revoke connector access.

## Files and terminal

- Notes are persisted app documents, not general storage.
- `SINet terminal --local` later starts a **local PTY** with ordinary OS permissions and a clear warning, never called a sandbox.
- `SINet usb list/copy` later enumerates mountpoints through a native local bridge and copies only explicitly named files after preview and confirmation. No automatic USB read, exfiltration, or autorun.
- `SINet terminal --cloud` later requires hardened isolated per-session compute. Never expose an API host shell or mount server secrets.

## Email and Web3 fidelity

`handle@owned-domain.tld` is an email alias requiring a controlled domain and mail-routing service. ENS `.eth` names can link cryptocurrency addresses and conventional email contact metadata; they are not inherently SMTP inboxes. A W3C DID is a separate identity primitive, not a replacement for verifying email or ENS name ownership.

References: https://modelcontextprotocol.io/specification/ ; https://docs.ens.domains/ensip/5/ ; https://www.w3.org/TR/did-core/
