# Product and system personas

## User personas

| Persona | Motivation | First-win moment | Barrier |
|---|---|---|---|
| Social explorer | Claim recognizable identity and meet people | Public handle + joined room | CLI anxiety |
| Terminal newcomer | Learn practical commands through guided action | Install/use first connector | Unknown terminology |
| Connector creator | Publish work and gain reputation | Connector reviewed, listed, installed | Publishing friction |
| Power user | Centralize tools and automate repetitive workflows | Permissioned agent finishes useful task | Trust and speed |
| Business operator (P1) | Branded namespace and delegated workflows | Separate organization profile and email | Credential/billing management |

## User-interface voice

Concise, useful, curious, welcoming. Never shame newcomers. Use simple verbs: `claim`, `find`, `join`, `plug`, `ask`, `note`, `settings`. Teach with one actionable hint at a time. No fake urgency or dark patterns.

## Runtime AI persona

**Personal Agent**: task-oriented assistant serving exactly one account's instructions within explicit connector grants. It can explain available tools, request consent, and report results or limitations, but cannot silently widen access. Canonical policy: `prompts/CLOUD_AGENT_SYSTEM.md`.

## Development review roles

Architect (minimal design), implementer (working slices), security reviewer (invariants), tester (evidence), UX editor (clarity). Definitions and output contract in `AGENTS.md`.
