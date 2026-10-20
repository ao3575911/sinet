# npm CLI distribution

Package: `@ao3575911/sinet`, version `0.1.1`.

Published on 20 October 2026 at https://www.npmjs.com/package/@ao3575911/sinet.
Fresh Linux x64 installation from the public registry passed with a separate cache
and installation scripts disabled. Both the installed `SINet version` command and
`npm exec --package=@ao3575911/sinet@0.1.1 -- SINet version` returned `0.1.1`.
Local tarball installation also passed for both command aliases. Packaging rejects
a modified release archive before writing output and refuses version changes
without updated release pins. Native macOS and Windows execution remains untested.

## Package design

`npm/` contains the package metadata and Node launcher. The distribution bundles
Linux x64, macOS arm64 and Windows x64 Go binaries from GitHub release v0.1.1.
There are no npm dependencies, installation scripts or runtime binary downloads.
The package includes the Apache license and third-party notices. It requires a
separately running SINet API; the server and development stack are not included.

`scripts/package_npm.py` reads the three CLI release archives from
`work/npm-release/`, verifies the pinned archive SHA-256 values and extracts only
the named binaries into `work/npm-package/`. It checks the repository and npm
versions against the pinned release. Updating versions requires reviewing new
archives and checksums. Checksums establish byte integrity; independent GitHub
provenance verification follows [the release guide](RELEASE.md). That verification
was not repeated in this environment because the GitHub API returned `Forbidden`.

```sh
python3 scripts/package_npm.py
npm pack ./work/npm-package --pack-destination ./work
npm install -g --prefix ./work/npm-local ./work/ao3575911-sinet-0.1.1.tgz --ignore-scripts
./work/npm-local/bin/SINet version
```

Consumers can use:

```sh
npm install -g @ao3575911/sinet@0.1.1
SINet version
npx --yes @ao3575911/sinet@0.1.1 version
```

Registry publication requires owner-controlled credentials. Never put tokens in
tracked files. The temporary npm authentication configuration used for this
publish request was removed after the command completed.

## Signed delivery

The original release signing private key was unavailable in this environment.
This delivery uses a new SSH signing key with fingerprint
`SHA256:qgf5+U/R3VF2oET0SVMKwoYl2gz8SxQTRX0CryYhj0E`; its public key is
`.github/release/npm-delivery-signing-key.pub`. Local Git signing is enabled and
uses ignored `work/signing/` material. The existing release trust list and original
release public key are preserved. This key is not authorized for release tags.
GitHub's Verified badge requires registering the new public key as an SSH signing
key in the maintainer account. No private signing key is committed or published.
