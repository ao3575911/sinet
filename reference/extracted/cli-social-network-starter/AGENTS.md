# Repository instructions for Codex

## Mission

Implement `docs/PRD.md` as a secure, runnable, lean social terminal platform. Prioritize a verifiable vertical slice over superficial feature count.

## Precedence

1. Security invariants in `docs/SECURITY.md`.
2. Product scope and acceptance in `docs/PRD.md`.
3. Design decisions in `docs/ADR/`.
4. API, architecture, UX, test plans.
5. This file and `prompts/CODEX_BUILD.md`.

## Working method

- Inspect repo and record observed state before editing.
- Write a brief implementation plan tied to acceptance tests, then implement in small increments.
- Do not invent functioning integrations, providers, payments, sandboxes, or email deliveries.
- Make every feature demonstrated with commands and tests. Describe blocked external services accurately.
- Avoid unnecessary dependencies, abstraction layers, comments, stubs, and generated noise.
- Never persist secrets in source, fixtures, logs, or commits.
- Use versioned migrations, bounded resources, typed interfaces, and explicit failures.
- Keep docs and command help synchronized with actual behavior.
- Run formatter, static checks, security checks available locally, and tests before completion.
- Never claim that uploaded arbitrary code is safely sandboxed without actually establishing and testing that isolation.
- Do not access local files or attached USBs except through explicit, scoped user action.
- Mark model-generated decisions as `ASSUMPTION:` in ADRs where not grounded in source.

## Agent roles

- **Architect:** define minimal contracts and record architecture decisions.
- **Builder:** implement complete vertical slices.
- **Security reviewer:** challenge identity, tenant isolation, tool grants, local I/O, SSRF, prompt injection.
- **Test engineer:** turn user journeys and abuse cases into automated checks.
- **Product editor:** keep command naming, onboarding, and documentation accessible to non-developers.

These are review lenses, not claims of separate autonomous agents.

## Completion report

Return only: working features, exact run commands, passing/failing tests, security limitations, and next implementation milestone.
