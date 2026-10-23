# SINet 0.2.0-rc.1 preparation

Source candidate only; no RC release/tag/package is published. Published downloads and
npm/PyPI stay 0.1.1. [Candidate notes](releases/0.2.0-rc.1.md) cover wallet ownership,
Gmail drafts and the completed local GitHub integration; [validation](releases/0.2.0-rc.1-validation.md)
separates passing local checks from the blocked live gateway gate. Follow [versioning](VERSIONING.md)
for a later reviewed release; do not move historical tags or overwrite 0.1.1 packages.

# SINet 0.1.1 prototype

Portable CLI packages: Linux amd64, macOS arm64 and Windows amd64. A separate Linux
amd64 API server package includes the example environment and operations guide. Go embeds
migrations, schemas and agent policy; no plugin executables are loaded. Packaging uses Go
and Python's standard library with no additional runtime libraries.

Download from https://github.com/ao3575911/sinet/releases/tag/v0.1.1.
Before extracting or running an archive, verify its GitHub build provenance using a current
GitHub CLI (authenticate with `gh auth login` if required):

```sh
gh attestation verify SINet-0.1.1-linux-amd64.tar.gz --repo ao3575911/sinet --signer-workflow ao3575911/sinet/.github/workflows/release.yml --source-ref refs/tags/v0.1.1
gh attestation verify SHA256SUMS --repo ao3575911/sinet --signer-workflow ao3575911/sinet/.github/workflows/release.yml --source-ref refs/tags/v0.1.1
sha256sum -c SHA256SUMS --ignore-missing
tar -xzf SINet-0.1.1-linux-amd64.tar.gz
./SINet-0.1.1-linux-amd64/SINet version
```

Replace the archive name for your platform; Windows uses ZIP. Checksums detect corruption;
verified OIDC/Sigstore attestations bind the bytes to the signed-tag build workflow. The
maintainer SSH key signs the annotated source tag and release commit; it never enters CI.
BUILDINFO.json records source commit, compiler and build flags. Archives omit build paths,
use fixed metadata and contain no credentials. They are not Apple-notarized or Windows
Authenticode-signed applications. The published 0.1.1 bytes predate native credential
storage. Current source has [native runtime and credential validation](NATIVE_CREDENTIALS.md);
new signed releases must pass that workflow before publication.

The CLI requires a running SINet API. To run the local development stack from source:

```sh
git clone https://github.com/ao3575911/sinet.git
cd sinet
git checkout v0.1.1
docker compose up -d --build
# Run the extracted CLI from its download directory in another terminal:
/path/to/SINet-0.1.1-linux-amd64/SINet
```

The API defaults to http://127.0.0.1:8080 and development email is visible in Mailpit at
http://127.0.0.1:8025. Server-only users must provision PostgreSQL and SMTP and configure
the environment explicitly; the binary does not read .env automatically. See
https://github.com/ao3575911/sinet/blob/main/docs/OPERATIONS.md.

This release implements identity, profiles, directory/lobby, moderation, private notes,
settings, a per-account agent and a permission-controlled connector registry with safe
echo. A model requires explicit provider configuration; no real remote connector or
public hosted service is bundled. Independent security review, production operations,
real connector validation, load testing and native desktop testing remain gates #2–#6.
