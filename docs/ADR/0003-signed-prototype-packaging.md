# ADR 0003: portable packages and signed build provenance

Accepted 2026-10-20. Publish 0.1.0 as a GitHub prerelease, with three portable CLI archives,
a Linux backend archive, BUILDINFO.json and SHA256SUMS. Use existing Go dependencies and
Python standard-library packaging; defer package-manager registries and container registry
publishing until their maintenance and trust policies are defined.

Require a trusted SSH-signed annotated version tag, a signed commit reachable from main,
SemVer/changelog agreement, integration tests and vulnerability scanning before packaging.
GitHub OIDC/Sigstore signs provenance for every asset, including the checksum manifest;
verify that provenance before publishing a draft. Check every uploaded asset's GitHub
SHA-256 digest before making the prerelease public. Never overwrite a published release.
No private maintainer signing key is stored in Actions. This does not replace platform
notarization, native platform tests, independent security review or production release gates.
