# Changelog

Versions follow Semantic Versioning 2.0.0. This project is in its 0.x prototype phase.

## [0.2.0-rc.1] - 2026-10-22

Release candidate source only; no release, tag or registry package published.

- Fixed public GitHub CLI/TUI journey with both-recipient disclosure, reviewed pinned
  installation, operation grant, one-use invocation consent, validated facts and audit.
- Gateway output is revalidated at the SINet API and CLI; bounded fixed failure codes.
- Counted two-account authorization, consent/revoke concurrency and failure fixtures.
- Upgrade build/CI toolchains to Go 1.27.3 after a reachable Go 1.27.2 x509 advisory.
- Hardened gateway container/healthcheck and operator deployment/rollback instructions.
- Live gateway/GitHub journey remains blocked on owned HTTPS hosting and operator review.

- Ethereum wallet ownership linking with one-time SIWE-format personal_sign proofs,
  a browser signing helper, private links, revocation and account export/deletion.
- Private email drafts and an explicitly approved Gmail compose/inbox handoff.
- Dashboard navigation and accurate setup states for wallet, email, business,
  external identity, files/devices, terminal/cloud and mobile.
- Account-isolation, replay, expiry, concurrent-proof, draft and browser fixture tests.
- No wallet custody/transfers, automatic Gmail sync or platform email provisioning.

## [0.1.1] - 2026-10-20

### Fixed

- Restore annotated tag objects after Actions checkout before checking signatures, and
  require the checked-out commit to match the signed tag. The 0.1.0 tag remains immutable;
  its workflow failed closed before any GitHub release was published.
- Include Go runtime and dependency license notices in binary distributions.

## [0.1.0] - 2026-10-20

### Added

- Go SINet CLI, keyboard onboarding/dashboard and PostgreSQL API with versioned migrations.
- Real development email OTP, hashed revocable sessions, unique handles and public profiles.
- Directory, membership-protected lobby, reports/blocks, persistent notes/settings and export/deletion.
- Private agent resource, optional disclosure-approved model chat and honest no-model mode.
- Reviewed connector manifests, pinned installs, operation grants/revocation, one-use consent,
  bounded HTTPS egress, audit and deterministic echo integration.
- Quotas, aggregate metrics, moderation/retention commands and operating/recovery guidance.
- Signed delivery, labelled issues/PRs/discussions, wiki, documentation site and CI workflows.
- Local semantic version command and verification of signed annotated release tags.

### Fixed

- PostgreSQL initialization readiness race with TCP health checks and bounded startup retries.
- OTP/handle concurrency and tenant/tool permission boundaries covered by integration tests.
- Coalesced TUI key handling and workflow health-command/ShellCheck issues found in validation.
- Windows dependency advisory GO-2026-5024 by updating x/sys to v0.49.0.

### Limitations

No production model/mail credentials or approved real remote connector are bundled. Native
macOS/Windows runtime UX, independent security review, sustained load and encrypted off-host
recovery qualification remain release gates. No untrusted uploaded code, host/USB shell,
active inbox, wallet or production-readiness certification is provided.
