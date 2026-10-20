---
layout: default
title: Architecture and boundaries
---

The host CLI talks to a Go REST API. PostgreSQL stores accounts, hashed authentication
records, profiles, agent state, room membership, notes, registry state, grants and audit.
Mail and model adapters are explicit external boundaries. The model has chat authority
only; tool execution passes through authenticated backend authorization.

Approved connector versions are immutable through the public API. No process runner,
arbitrary uploaded code or host/USB tool exists. Remote HTTPS execution pins validated
public IPs, checks the peer, denies redirects/proxies, and bounds arguments, time and output.

See [architecture](https://github.com/ao3575911/sinet/blob/main/docs/ARCHITECTURE.md) and
[recorded decisions](https://github.com/ao3575911/sinet/tree/main/docs/ADR).
