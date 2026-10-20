# Repository maintenance

## Labels and review

Type labels: bug, enhancement, documentation. Area labels: area:cli, area:identity,
area:connectors, area:operations, area:docs. Security/release labels: security, release-gate,
priority:high, status:blocked, prototype. Automation changes use automation. Release intent uses semver:major/minor/patch.

Release issues #2–#6 contain verifiable public-launch acceptance. #7 records the fixed
PostgreSQL startup race. #8 tracks documentation/community publication. PR #1 is labelled,
links evidence and is ready for review; the owner explicitly authorized delivery to main.
Future PR area labels are applied by a metadata-only action that never checks out or runs
untrusted PR source. The label configuration is .github/labeler.yml.

## Actions

| Workflow | Trigger | Purpose |
|---|---|---|
| semantic-version | main, PR, version tags and manual | SemVer/changelog/CLI agreement; signed annotated tag and release commit verification against main trust list |
| validate | push and PR | Two automation tests; workflow lint; fmt/vet/module checks; real PostgreSQL race tests; cross-build; Mailpit CLI/PTY journeys; govulncheck |
| dependency-security | weekly, manual and scanner configuration changes | Current vulnerability database scan without requiring live application credentials |
| portable-binaries | main client changes, version tags and manual | Linux/macOS/Windows CLI artifacts and SHA256SUMS, retained 14 days; not an automatically signed production release |
| label-pull-requests | PR metadata events | Add area labels using trusted-base configuration; no PR code execution |
| community-setup | main configuration changes and manual | Create custom labels, apply issue/PR labels, create discussion starters once and label them |
| documentation-pages | main documentation changes and manual | Build Jekyll documentation and deploy GitHub Pages through the configured environment |

Community automation is idempotent and tested against repeated runs. It reports disabled
Discussions instead of pretending to publish them. It requests only contents-read,
issues/pull-requests/discussions write permissions. Pages deployment uses pages-write and
OIDC; neither workflow reads application mail/model secrets or the local signing key.
Owner enabled Discussions, initialized the wiki and activated Pages for this delivery.

## Documentation

Canonical detailed contracts: docs/*.md. Community guides: docs/wiki/*.md. Published wiki:
https://github.com/ao3575911/sinet/wiki . Documentation site source: site/; deployment URL:
https://ao3575911.github.io/sinet/ . Publication is verified through Actions before being
claimed complete in the final evidence ledger.

To update the wiki with a provisioned signing identity, run `python3 scripts/publish_wiki.py`.
It uses the root repository's signing configuration, signs the wiki commit, verifies the
signature, pushes over the configured Git transport and checks the remote SHA. It only
copies versioned Markdown guides and never reads application credentials. Its checkout
and signing material remain under ignored work/.

## Signed delivery

All authored source/wiki commits require signatures. Verify locally and query GitHub's
commit verification before reporting a successful main push. Compare `git rev-parse HEAD`
with `git ls-remote origin refs/heads/main`. Do not force-update shared history. GitHub
transports signed commits over authenticated HTTPS; it does not support push certificates.

Keep the dedicated signing key outside source and securely manage its lifecycle after
this environment ends. No production private keys or signing credentials belong in
Actions artifacts, Pages, wiki, issue bodies or Discussions.


## Distribution status

Published GitHub prototype: https://github.com/ao3575911/sinet/releases/tag/v0.1.1 .
Three CLI targets and Linux server archives have verified build provenance and checksums.
The source tag and release/main commits are signed. npm/npx and PyPI/pip packages and a
Marketplace Action are not listed. Issue #14 records registry ownership/authentication,
installer validation and Action listing prerequisites. The Go application is not published
as a JavaScript or Python library. Marketplace lists Actions/apps; no listing is claimed.
