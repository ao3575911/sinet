# ADR 0004: ship the Go CLI through npm and PyPI

Accepted 2026-10-20 for distribution preparation. Use npm package `@ao3575911/sinet` and
PyPI project `sinet-cli`, conditional on owner-controlled registry accounts. Public name
availability checks are not reservations or proof of account ownership.

Bundle existing, exact-source-attested Go release binaries. npm ships three binaries with
a Node standard-library launcher; PyPI ships one platform wheel per binary with a Python
standard-library launcher. There are no install scripts, runtime downloads, additional
runtime libraries or arbitrary executable discovery through PATH. Validate the binary hash
and regular-file type before invocation. Preserve licenses and original Go BUILDINFO.

Use npm pack and a small standard-library wheel packager, with wheel RECORD hashes and
executable modes. Validate metadata with pinned Twine and real npm/npx/pip installations.
The current registry version policy accepts final x.y.z versions; prerelease SemVer-to-PEP
440 mapping must be defined before publishing a prerelease identifier to both registries.

CI verifies trusted signed source and original release attestations, attests candidate
packages, and gates publication on Linux/Windows native install and tamper tests. macOS
remains cross-build-only evidence. Use npm provenance and PyPI Trusted Publishing; the
initial npm package may need a narrowly scoped bootstrap token. Never store maintainer
signing keys in Actions. Tokens are provisioned out of chat by the account owner.

The owner requested registry publication before merging main. Build/test on a signed
preparation branch; explicit signed `npm-vVERSION` / `pypi-vVERSION` intent tags can run the
corresponding publishing job before the new workflow is available on the default branch.
Verify registry hashes/provenance and clean installs before merging. Never overwrite an
existing registry version. Defer Marketplace until a tested Action and owner listing exist.
