---
layout: default
title: Security and privacy
---

Email, sessions, notes, agent history, settings and audit are private account data.
Public profiles show selected public fields. Making a profile private removes its public
JSON/HTML/directory view. Blocking filters room/directory interaction; deliberately public
profile pages remain accessible anonymously.

OTP codes expire, are stored as keyed hashes and redeem once under concurrency. Session
tokens are hashed server-side and revocable. Local Unix credential files use mode 0600;
OS keychain integration and native Windows ACL verification remain release work.

Export is paginated; deletion removes live account/private rows and preserves detached
aggregate/audit history and handle tombstones as documented. Backups follow their expiry.

Do not post secrets or exploit details publicly. Read the
[security reporting policy](https://github.com/ao3575911/sinet/blob/main/SECURITY.md).
