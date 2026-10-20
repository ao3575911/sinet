---
layout: default
title: Versions and signed releases
---

The current prototype version is **0.1.1**. VERSION is authoritative; `SINet version` and
`SINet --version` report it locally without API access. CHANGELOG.md records dated changes.

Use SemVer 2.0.0. In 0.x, use a minor bump for features/breaking public changes and patch
for compatible fixes. Release tags are v plus VERSION. Prereleases such as 0.2.0-rc.1 are
supported. Do not move a published tag; issue a new version.

The semantic-version Action checks syntax, changelog and compiled CLI agreement. A tag
must be annotated and signed, match VERSION, point to a signed commit reachable from main,
and verify against the public trust list from main. Private signing keys never enter CI.
[Prototype release v0.1.1](https://github.com/ao3575911/sinet/releases/tag/v0.1.1) includes
three CLI archives, a Linux server archive, BUILDINFO.json and SHA256SUMS. Every asset has
verified GitHub OIDC/Sigstore build provenance; the annotated source tag is SSH-signed.
See [installation and verification](https://github.com/ao3575911/sinet/blob/main/docs/RELEASE.md).
npm/npx, PyPI/pip and Marketplace publication remain [tracked in issue #14](https://github.com/ao3575911/sinet/issues/14).

See [full versioning/release procedure](https://github.com/ao3575911/sinet/blob/main/docs/VERSIONING.md)
and [changelog](https://github.com/ao3575911/sinet/blob/main/CHANGELOG.md).
