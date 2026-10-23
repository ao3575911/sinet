# Native credentials and runtime validation

Issue [#6](https://github.com/ao3575911/sinet/issues/6) covers macOS arm64 and
Windows amd64. The `native-runtime` workflow runs on macOS 15 and Windows Server
2025. These are the tested platforms; other desktop versions are not implied.

## Session storage and migration

On macOS, tokens live in Keychain; on Windows, in the current user's Credential
Manager. `SINET_CONFIG` selects a small JSON metadata file containing the API origin
and a random credential reference. Store keys bind that reference to the absolute
config path and origin. Copying the file or editing its origin cannot rebind the
token. Secrets are not passed as process arguments or printed by SINet. The pinned
`go-keyring` adapter uses the macOS security tool's stdin and native Windows APIs.
There is no automatic plaintext fallback when the store is locked or unavailable.

Windows session files have an explicit, protected DACL granting only the current
user access. The loader checks the opened file's owner and ACL, including legacy
files. Unix files retain mode 0600. Symlinks, nonregular files, oversized files and
files replaced while opening are rejected. Keep parent directories private and
trusted; these controls do not defend against compromise of the logged-in OS user.
Linux retains the existing mode-0600 file storage; this change does not add a Linux
Secret Service dependency.

A valid legacy plaintext session migrates when loaded for its original API. The
store is written and read back before the file is atomically replaced. A failed
save leaves the previous session active. A changed `SINET_API` never causes migration
to the new origin. Broad Windows ACLs fail closed: remove the old session file and
sign in again. Moving a native config file also requires a fresh sign-in. Old
binaries reject token-free metadata; sign out before downgrading. Logout revokes
the server session and removes both the local reference and OS credential.

An interrupted save can leave an unused OS-store item. A failed cleanup does not
make a revoked server token valid. Such orphaned items can be removed from the OS
credential manager under service `SINet session v1`. SINet never enumerates or
deletes other applications' credentials.

## Reproducing native validation

Install Go 1.27.3, Python 3.12 and PostgreSQL tools on PATH. Put Mailpit 1.27.10 in
`bin/mailpit` (`bin/mailpit.exe` on Windows). Windows terminal testing requires
`pywinpty==3.0.2` and permission to create symlinks. macOS CI creates an unlocked,
disposable keychain. Then run:

```sh
python tests/native_runtime.py
```

This creates an isolated loopback-only database and SMTP sink, runs race tests
against PostgreSQL, builds the release-style CGO-disabled CLI, and exercises real
email OTP onboarding, private notes, grants/revocation, logout/relogin and deletion.
The terminal journey uses POSIX PTY on macOS and ConPTY on Windows for guided signup,
hidden OTP input, note editing, arrow keys, numbered navigation and quit. Native
credential tests use the real OS store, check migration and missing/wrong-origin
references, replace/delete stored tokens, and reject symlinks. The Windows test
also deliberately grants Everyone access and verifies rejection. Fixture accounts
and temporary schemas must be gone before the run succeeds.

Only after all checks pass does the workflow upload the tested binary,
`SHA256SUMS`, and `VALIDATION.json` with exact commit/compiler/platform evidence.
Artifacts are retained for 30 days and are downloadable from the workflow run.
They are tested candidates, not notarized or Authenticode-signed releases. The
signed release workflow requires this native workflow before it can publish.
Independent security review and production deployment gates remain separate.

## Dedicated release-signing key lifecycle

The release trust root is `.github/release/allowed_signers`. A release requires a
registered maintainer signing identity, a signed commit reachable from main and
a signed annotated tag. CI receives no maintainer private key; OIDC provenance
binds release assets to the workflow and source. Verify both provenance and hashes
as described in [release verification](RELEASE.md).

Keep the dedicated private release key outside the checkout and CI, protected by
a passphrase or hardware-backed storage, with an encrypted offline recovery copy
controlled by the owner. Restrict access to release maintainers. Before rotation,
register the successor public key as a GitHub signing key, update the reviewed
allowlist while the old key is still valid, and verify a signed test commit. Use
the successor for subsequent tags, remove the old active authorization, and retain
historical public keys and provenance for verification. Never rewrite old tags or
replace published assets to rotate a key.

If a key is lost or suspected compromised, pause release publication, revoke its
GitHub signing registration and active allowlist entry, inspect releases made in
the exposure window, and publish a dated incident notice with affected versions
and replacement verification instructions. Restore signing through a separately
reviewed successor key. Revoke stale bootstrap registry tokens after trusted
publishing is configured. Key generation, custody transfer and rotation require
the owner's actual key-management steps; documentation is not evidence they occurred.

## Validation evidence

Local Linux validation on 2026-10-23 passed `make check integration cross` with
PostgreSQL 18.6, the full two-account SMTP/CLI smoke journey (including logout and
fresh sign-in), and the PTY onboarding/editor journey. All fixture accounts and
test schemas were removed. Seventeen Python packaging tests, two Node automation
tests and actionlint passed. `govulncheck` found zero reachable vulnerabilities
in Linux, macOS arm64 and Windows amd64 graphs; the existing unused OpenPGP module
advisory remains unrelated to imported code.

Final PR review on 2026-10-24 (Australia/Perth) reproduced the Windows acceptance
failure: Python text-mode stdin changed the note fixture's LF into CRLF. The
harness now sends exact UTF-8 bytes and uses explicit native executable names.
The full local database/race suite, cross-builds, CLI/SMTP and PTY journeys,
17 Python tests, two Node tests and actionlint passed again; fixture accounts
and test schemas were confirmed absent. No submitted reviews or unresolved
review threads existed at review time.

[Native run 40037079258](https://github.com/ao3575911/sinet/actions/runs/40037079258)
passed both macOS 15 arm64 and Windows Server 2025 amd64, including real credential
stores, PostgreSQL/race tests, CLI/SMTP, native terminal journeys and fixture cleanup.
The downloaded binaries were independently hashed against both SHA256SUMS and
VALIDATION.json. That run tested PR head `6b9b45fd122c9b9dacabd64b3ffa0ac5547438dc`
through GitHub's merge ref `cdb9637f92280ca99c98af275409ef9bdc2def84` before history
consolidation; the final consolidated commit must pass the PR checks again.

| Tested artifact | SHA-256 |
| --- | --- |
| [macOS arm64](https://github.com/ao3575911/sinet/actions/runs/40037079258/artifacts/12537866348) | `14281304855dd301f9a2a6a24bcd6212315ffbefa9e280eea3ee6462d03e5b77` |
| [Windows amd64](https://github.com/ao3575911/sinet/actions/runs/40037079258/artifacts/12537662201) | `c886d52c79803e076900c3a6f068db6b0d1363693690197cf3f0ddd87edbc7f4` |

Both artifacts record Go 1.27.3. Downloads expire after the workflow's 30-day
retention period. Final commit checks and any newer artifact links are available
on [PR #22](https://github.com/ao3575911/sinet/pull/22).
