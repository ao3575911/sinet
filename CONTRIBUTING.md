# Contributing to SINet

SINet is a production-oriented prototype. Read README.md, AGENTS.md, docs/PRD.md,
docs/SECURITY.md and docs/ADR/ before proposing changes. Preserve reference/ originals.

1. Open or select a labelled issue with a concrete outcome and acceptance checks.
2. Implement a small complete slice. Keep the Go dependency set lean and permissions
   default-denied. Document assumptions and architecture decisions.
3. Run `make check integration`; use `make smoke smoke-tui` for CLI/API flows and
   `make cross` for portable client changes. Database tests must not be skipped.
4. Update command help, documentation and evidence. State external service limitations.
5. Sign every commit using a key registered with GitHub, verify with `git verify-commit`,
   and open a labelled PR. Never commit tokens, private keys, OTPs or runtime exports.

There is no arbitrary plugin executor, general hosted shell, cloud host/USB bridge or
production email inbox. A PR must not silently introduce those capabilities. New connector
permissions, data exports and irreversible effects need explicit authorization design.

Use bug/enhancement/documentation and area labels. `release-gate` tracks work required
before a public launch; `status:blocked` means a named external dependency is missing.
Do not mark work complete from screenshots or compilation alone. CI artifacts are build
outputs, not automatically signed production releases.
