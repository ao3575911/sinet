# GitHub connector completion sprint — 22 October 2026

Base: origin/main d017f7b0614f31fcace58ffc67fab38c2b73f3d3. Original checkout clean;
root-owned files preserved unchanged. Work lives in /tmp/sinet-connector-sprint on
codex/github-connector-completion. Origin has only main; historical feature branches
are already incorporated. All five existing tags are preserved.

## Initial state

- [x] Fixed GitHub gateway, closed schemas, synthetic adversarial fixtures,
  safe HTTPS transport, registry review, grants, one-use consent and audit inspected.
- [x] Unreleased wallet ownership/Gmail draft implementation present; preserve it.
- [x] Independently verify published npm/PyPI 0.1.1 registry metadata.
- [x] Issues #2–#6 and #14 and prior PRs inspected; these remain independent gates.
- [x] Complete disclosure naming gateway and GitHub, fixed-input CLI/TUI flow,
  API-side result validation and bounded useful failures.
- [x] Counted two-account integration: success and denial/isolation/revocation cases.
- [x] Hardened gateway deployment, health/monitoring and rollback preparation.
- [x] Run application checks, SMTP/PTY/browser journeys; record precise blockers.
- [ ] Prepare 0.2.0-rc.1 notes/version, review diff, sign/verify, push branch and open PR.

## Acceptance criteria

An activated account inspects and installs reviewed github-public@1.0.0, grants only
repository_summary, reviews exact owner/repository and both recipients, approves once,
invokes through the deterministic personal agent, reads validated public facts/audit,
and revokes. Missing authority makes zero counted upstream requests. Consent stays
account/version/operation/input bound, five-minute and single-use. A concurrent revoke
waits for an already authorized call holding the account transaction lock; after revoke
returns, new calls cannot execute. No model, email sending or wallet transaction required.
Adversarial fixtures stay isolated from live evidence; safe egress remains mandatory.

Validation details and corrected initial failures: [candidate ledger](releases/0.2.0-rc.1-validation.md).

## External gates

No owned public gateway hostname, ownership evidence, hosting access or operator review
is configured. Owner delegated deployment design; choose a hardened single-host container
behind operator-managed TLS ingress. Do not invent ownership, approve placeholders or
weaken public-IP checks. Live gateway/GitHub acceptance stays blocked until provisioned.
Docker socket is unavailable to this user; use isolated user-space development services
if available. Original workspace sandbox fails before execution; approved commands run
outside it, with no network-policy bypass. No release/tag/package publication allowed.
