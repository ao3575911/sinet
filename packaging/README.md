# SINet CLI installers

Distribution names: **@ao3575911/sinet** (npm/npx) and **sinet-cli** (PyPI/pip).
These are thin launchers for the existing Go application, not JavaScript/Python APIs.
Status: [sinet-cli 0.1.1 is published on PyPI](https://pypi.org/project/sinet-cli/0.1.1/)
and verified with public file hashes and a fresh pip install. npm publishing is blocked
by the registry's E404 permission/scope response for @ao3575911/sinet; no npm publication
is claimed. The signed preparation branch remains unmerged until npm is verified.

The npm package bundles Linux x64, macOS arm64 and Windows x64 binaries. Python uses
separate wheels for each target. macOS wheels require macOS 14+; the Linux wheel uses
the manylinux_2_28_x86_64 platform tag and a static CGO-disabled executable. The Windows
wheel requires x64 Windows supported by Go 1.27. Native macOS execution is not yet
validated. No ARM Linux or Intel macOS binary is bundled. Requirements: Node.js 22+ or
Python 3.9+. Registry packaging currently accepts final x.y.z versions; prerelease
SemVer-to-Python-version mapping is not implemented. A running SINet API is still required for sign-in and network commands.

There are no post-install hooks, runtime dependencies or runtime downloads. Each launcher
checks its bundled binary's SHA-256 and refuses a changed binary or a symlink. Registry
package provenance authenticates the distribution; these local checks detect corruption
but cannot protect against an attacker who can replace the launcher and manifest together.
All Go binary bytes originate in the verified GitHub release, with build information and
third-party licenses retained. Do not describe them as platform-notarized executables.

The pip commands below are verified against the public registry. npm/npx commands are
publication targets awaiting scope access:

```sh
npx @ao3575911/sinet@0.1.1 version
npm install --global @ao3575911/sinet@0.1.1
SINet version
python3 -m pip install sinet-cli==0.1.1
SINet version
```

Use a Python virtual environment where required by the operating system. No npm/npx
publication is claimed until the upload and fresh registry installation pass.

## Build and verify candidates

The registry-distribution Action downloads the public release, verifies its annotated
source tag and every asset's exact-source GitHub attestation, checks SHA256SUMS, then builds
the npm tarball and three wheels. It uses npm's packer and Python's standard library for
the wheel archive/metadata/RECORD; no custom installation code executes during pip install.
It signs provenance for the candidates and runs real npm/npx/pip installs and tamper-denial
tests on Linux and Windows. Both platforms passed in
[registry CI 39366222801](https://github.com/ao3575911/sinet/actions/runs/39366222801).
Native macOS remains a release gate.

For an already verified local release directory:

```sh
python3 scripts/package_registries.py --source work/registry-source --output work/registry-packages
python3 tests/registry_install.py --packages work/registry-packages
```

The source directory must contain all six original release assets. Verify their provenance
as documented in docs/RELEASE.md before running the local builder. The builder checks
checksums and source metadata again; local checksum comparison alone is not authentication.

## Owner setup required before publication

Before the first publication on 20 October 2026, public lookups found neither chosen name registered. Availability does
not reserve a name or establish ownership. Use an owner-controlled npm account/scope
`ao3575911` and an owner-controlled PyPI project `sinet-cli`; change the scope if the owner
cannot control it. Never publish using somebody else's namespace or paste tokens in chat.

For npm's first publish, provision package/scope publishing access using the environment's
npm login or a GitHub Actions secret named NPM_TOKEN. The publish workflow uses a modern
npm with --provenance. Once the package exists, configure npm Trusted Publishing for
ao3575911/sinet, workflow registry.yml, environment npm, then remove the bootstrap token.
For PyPI, configure a pending Trusted Publisher for sinet-cli, owner ao3575911, repository
sinet, workflow registry.yml, environment pypi. No PyPI token is required with that setup.
The registry account owner must perform those external account settings. The PyPI
configuration has now completed a successful trusted publish. The npm token reached
publication but the registry rejected package creation in @ao3575911; the owner must
confirm scope ownership and token/publishing permissions before retrying that job.

After registry setup and passing candidate checks, publish from the signed preparation
branch before merging main by creating signed intent tags:

```sh
git tag -s npm-v0.1.1 -m 'Publish tested SINet 0.1.1 npm package'
git verify-tag npm-v0.1.1
git push origin npm-v0.1.1
git tag -s pypi-v0.1.1 -m 'Publish tested SINet 0.1.1 Python wheels'
git verify-tag pypi-v0.1.1
git push origin pypi-v0.1.1
```

Do not push these tags until the corresponding registry access is configured. After the
workflow exists on main, manual registry-distribution runs with explicit publish inputs
are also available. New workflow_dispatch files on a feature branch alone cannot be
manually dispatched; signed intent tags make the owner's publish-before-main sequence
possible without weakening source verification. Publishing jobs require
both native install jobs to pass. PyPI requires Trusted Publishing; npm accepts configured
Trusted Publishing or the bootstrap secret. Versions are immutable: never replace published
package bytes. A preflight rejects conflicting existing npm/PyPI hashes; identical complete
versions can be verified again without uploading. After publication, the workflow compares
public registry metadata with candidate hashes, audits npm signatures/provenance, and runs
fresh npm/npx or pip installations before merging the signed branch to main.

If an intent fails before publication and needs a code fix, retain its immutable signed
tag and create a new signed intent such as `npm-v0.1.1-retry.1`. Retry counters are positive
integers; the channel and package version remain exact. This changes the publication
attempt, never an existing package version. The publisher uses an explicit `./packages/`
archive path so npm cannot interpret it as a GitHub repository shorthand.

Marketplace is deferred until after npm/PyPI. It needs a useful tested Setup SINet Action
and the owner's Marketplace listing step; the backend repository itself is not a listing.
