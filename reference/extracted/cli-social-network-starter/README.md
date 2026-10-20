# project-name

**Your username. Your agent. Your terminal. Your network.**

A Unix-inspired social CLI where people claim an identity, meet others, and extend a personal cloud agent with permissioned connectors.

**Status:** pre-development specification. This package contains contracts and instructions, not a finished application. Replace `project-name` and the placeholder domain before public release.

## North star

Make the username desirable, the terminal approachable, and the connectors indispensable.

## Product contract

- One executable enters the network: `project-name`.
- Email verification and unique handle claiming precede account activation.
- Every account has a public profile, private agent identity, and platform email alias (configured mail service required).
- Users can discover profiles, join rooms, write notes, and install curated connectors.
- Connectors are the central differentiator: browse, inspect permissions, install, use, and submit a validated manifest. Curated built-ins and reviewed remote HTTPS connectors can run; uploaded code cannot.
- Local machine access, AI tool calls, USB operations, and remote workspaces are distinct permission boundaries.
- Business identities, crypto wallets, ENS, cloud shell, and mobile are staged expansions, not fictitious MVP functionality.

## Starter files

- `docs/CONCEPT.md`: polished product narrative
- `AGENTS.md`: binding Codex/repository instructions
- `docs/PRD.md`: product boundaries and acceptance
- `docs/ARCHITECTURE.md`: chosen technology and service boundaries
- `docs/SECURITY.md`: mandatory trust boundaries
- `docs/PERSONAS.md`: user and agent personas
- `docs/UX.md`: journeys and CLI commands
- `docs/ROADMAP.md`: milestone gates
- `docs/API.md`: service contract
- `docs/TESTS.md`: test/quality gates
- `docs/ADR/0001-runtime-and-scope.md`: key decisions
- `prompts/CODEX_BUILD.md`: copy into Codex
- `prompts/CLOUD_AGENT_SYSTEM.md`: runtime agent policy
- `contracts/connector.schema.json`: declarative connector contract
- `connectors/examples/echo.json`: safe demonstration plugin
- `db/schema.sql`: initial relational model, subject to migration review
- `.env.example`: configuration inventory, no credentials

## First delivery

A working end-to-end vertical slice: start CLI → verify email → claim unique handle → see profile → browse/install example connector → use it through the personal agent → join a room. Unavailable third-party AI must not be disguised as a functioning real model.
