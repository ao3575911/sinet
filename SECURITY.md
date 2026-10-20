# Security policy

SINet is a prototype; no version is yet certified for public production deployment.
The current development branch receives fixes. Security invariants and threat boundaries
are in docs/SECURITY.md; test and operational evidence is in docs/VALIDATION.md.

Do not disclose credentials, account data or exploitable vulnerability details in public
issues or Discussions. Use GitHub's private vulnerability reporting link if the repository
has it enabled: https://github.com/ao3575911/sinet/security/advisories/new . If unavailable,
request a private reporting channel from the maintainer without including exploit details.
No unverified security email address or response-time guarantee is advertised.

Reports should describe the affected commit, safe reproduction, expected boundary and
observed impact. Never test against other people's accounts or unapproved remote services.
Maintainers should triage privately, preserve minimal evidence, fix with regression tests,
review dependencies, and publish coordinated advisories when appropriate.

Supported development controls include hashed one-use OTPs/sessions, tenant filtering,
immutable connector review, grants, one-use consent, bounded egress and quotas. The model
cannot grant or execute tools itself. No uploaded code or host shell is provided. Production
TLS, credential rotation, monitoring and independent review remain deployment gates.
