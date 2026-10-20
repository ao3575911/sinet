# PRD: Unix-Inspired Social CLI

## Vision

Make the terminal a welcoming social destination with a memorable handle, a useful personal AI agent, and a connector marketplace. Hero action: **install a connector, make the agent do something useful, and share the result**.

## Users

- Social-native explorer: wants community and distinctive identity without learning Unix.
- Terminal newcomer: wants discoverable actions, friendly help, reversible commands.
- Builder/creator: wants to publish a useful connector and gain attribution.
- Operator/business (later): wants separate branded identity and delegated tools.

## Primary experience

`SINet` opens a terminal UI. User verifies email, claims one globally unique normalized handle, lands in a discoverable profile, sees their private agent and platform email alias, joins a room, installs a curated connector, and executes a tool via agent approval. A profile URL/card can be shared outside the CLI.

## P0: implement for the first working MVP

1. CLI executable/TUI for Linux/macOS/Windows with keyboard navigation and plain noninteractive subcommands.
2. Email OTP sign-in: one-time short expiry code, replay denial, hashed stored code, delivery through configured provider; Mailpit in development.
3. Atomic unique handle claiming; normalization, reserved names, anti-impersonation, rate limits; public profile, shareable text card, and minimal web profile page with link preview.
4. Personal agent resource and chat interface: configurable OpenAI-compatible model endpoint/BYOK or plainly marked no-model mode; never fabricate responses from unavailable AI.
5. Public directory and a single community room, messages with pagination and abuse reports/blocks.
6. Lightweight per-user notes/editor; edits persist server-side; no general cloud drive/file explorer.
7. Connector registry: list/details, capability declaration, upload validated manifest, inspect permissions, install/uninstall, and call one built-in safe connector through the agent. Publisher-submitted manifests start in review state; vetted remote HTTPS connectors may be approved with verified ownership and locked-down outbound networking. Never execute uploaded code.
8. Real capability grants scoped to account + connector + operation. User approval for writes and external disclosures. Platform email alias is generated as an address/claim only when the platform controls the domain; show deliverability/provisioning state and never claim an active inbox without routing.
9. Account settings: privacy, theme, notification defaults, installed connectors, model usage.
10. Instrument basic funnels and usage with privacy-preserving aggregates, quota enforcement, account deletion.

## P1: after verified traction

- Platform-domain email forwarding aliases (if unavailable in MVP infrastructure, reserve/display provisioning status instead of pretending mail is deliverable); business profile + handle + dedicated agent + verified-domain email.
- More vetted connectors, secure OAuth, creator pages, ratings and install analytics.
- Local terminal PTY module and removable-drive copy via separate native local bridge, default off.
- Portable isolated cloud workspace using vetted container/VM infrastructure, separately billed.
- Public web profile viewer and a responsive mobile-friendly companion UI.

## P2: options, not promises

- Noncustodial optional wallets, ENS records/verified ownership, W3C DID-based contextual identity.
- Business commerce, escrow/payment integrations with relevant legal review.
- Richer social graphs and federated/distributed identity features.

## Explicit non-goals for P0

No unrestricted cloud shell, arbitrary uploaded plugin execution, agent execution on the user's host, full email inbox, E2E messaging claims, crypto custody, guaranteed .eth email delivery, token minting, or iOS-native terminal/USB access.

## Growth principles

Desirable first-come handle claims, transparent reserved-name rules, shareable profiles, beginner missions, community-led connector showcases, and opt-in streaks/achievements. No manufactured scarcity, spam incentives, infinite-scroll dark patterns, or unverifiable follower metrics.

## Success and kill metrics

- Activation: verified signup → claimed handle → first connector installed/used.
- 7-day retained users; % of users achieving first useful connector outcome.
- Trust: unauthorized tool calls = zero in test suite, abuse reports resolved, fair handle ownership.
- Costs: hosted agent spend and fraud rate per active user.

## Launch gate

A new user can complete the P0 journey from a clean machine and the suite proves account isolation, handle atomicity, OTP limits, plugin grants, room auth, and no host-file access by the agent.
