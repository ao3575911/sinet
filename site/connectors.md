---
layout: default
title: Connector marketplace and agent
---

Use `SINet plug list`, `plug info`, `plug install`, `plug grant`, `agent call`, `agent audit`
and `plug revoke`. Installation pins a reviewed version. A grant is account/connector/operation
scoped. Writes and all remote disclosures also need a fresh input-bound one-use consent.

`plug validate FILE` and `plug submit FILE` explicitly read a user-selected local manifest.
Submitted metadata enters review; it never executes uploaded code. Publisher attribution
must match the account's handle. Remote services require independent ownership and security
review. An operator approval reference is an attestation, not automatic ownership verification.

No real remote service is preapproved. Valuable irreversible remote operations need upstream
idempotency and reconciliation; do not retry uncertain side effects automatically. See
[operating contract](https://github.com/ao3575911/sinet/blob/main/docs/OPERATIONS.md).
