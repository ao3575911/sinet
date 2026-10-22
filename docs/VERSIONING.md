# Semantic versions and signed releases

VERSION is the source of truth. Source is **0.2.0-rc.2**; the published prototype and
npm/PyPI packages remain **0.1.1**, immutable. No RC tag or release has been published.
The CLI reports its compiled source version through `SINet version` and `SINet --version` without contacting the API. CHANGELOG.md
records dated release behavior and limitations. Never label a prototype as production.

Follow SemVer 2.0.0. During 0.x development, increment the minor component for features or
breaking public CLI/API/manifest changes; use patch increments for compatible fixes. After
1.0.0, breaking public changes increment major, compatible additions minor and fixes patch.
Prereleases use identifiers such as `0.2.0-rc.2`; build metadata does not change precedence.
Numeric identifiers cannot have leading zeroes. Release tags are `v` plus VERSION.

Prepare a release on a reviewed branch: update VERSION and the dated changelog entry,
run validation, build the CLI, and run:

```sh
python3 -m unittest discover -s scripts -p test_version.py -v
python3 scripts/check_version.py --binary ./bin/SINet
```

After the reviewed release commit is signed, verified and present on main, maintainers
create and verify the annotated tag using their provisioned signing key:

```sh
# Future reviewed publication only; not part of the current sprint.
git tag -s v0.2.0-rc.2 -m 'SINet 0.2.0-rc.2 — source release candidate'
git verify-tag v0.2.0-rc.2
git verify-commit 'v0.2.0-rc.2^{}'
git push origin v0.2.0-rc.2
```

The semantic-version Action validates syntax, dated changelog and compiled CLI agreement.
For tags, it rejects lightweight tags, mismatched versions, untrusted signatures, unsigned
release commits and commits outside main. Its SSH trust list is loaded from main's public
.github/release/allowed_signers. Key changes require maintainer review; public key fingerprint
is SHA256:DZbp/eHiq7txsnGklKuBtZ1XOotv927BpKkCIfS5ggw. The private key never enters Actions.

The signed-prototype-release Action verifies source signatures, runs integration tests and
vulnerability scanning, packages the CLI/server, and signs every asset's build provenance
using GitHub OIDC/Sigstore. It verifies those attestations and checks uploaded SHA-256 digests
before publishing a GitHub prerelease. See [installation and verification](RELEASE.md).
The portable-binaries Action also retains temporary CI artifacts. Neither workflow receives
the maintainer private key. Runtime/operational release gates remain issues #2–#6.
Never move an already published tag; publish a new SemVer instead.

A draft whose assets already have valid provenance can be finalized separately. A trusted
signed main commit records its exact tag/source SHA in `.github/release/publish.json`.
The finalize-verified-release Action verifies the intent commit, release tag/source,
asset provenance (including exact source digest), checksums and authenticated draft asset
digests before publishing. It neither rebuilds nor replaces assets. This recovery path
handles GitHub's published-only release-by-tag endpoint; draft verification uses the
maintainer-authenticated draft-inclusive release listing. Publisher regression tests
reject digest mismatches and prohibit uploads in existing-only mode.
