---
layout: default
title: Operations and recovery
---

The API runs nonroot with a read-only root filesystem, no capabilities and no host mounts.
Compose waits for PostgreSQL TCP readiness; the API has a bounded startup retry. Migrations
are versioned, transactional, advisory-locked and checksum-protected.

Production requires TLS ingress, verified database TLS, secure credentials and STARTTLS mail.
A configured model requires explicit disclosure approval and provider-side spending limits.
A controlled mail domain/routing is needed before an email alias can become active.

Operators can review connectors, resolve reports, inspect aggregate metrics and run retention
maintenance. Backup restoration has been tested into a separate database; encrypted off-host
backup rotation and sustained-load testing remain release gates.

Exact commands, quotas, retention and recovery procedures:
[OPERATIONS.md](https://github.com/ao3575911/sinet/blob/main/docs/OPERATIONS.md).
