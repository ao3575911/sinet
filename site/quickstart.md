---
layout: default
title: Quickstart
---

Install a current patched Go compiler (tested with Go 1.27.3) and Docker Compose.

```sh
git clone https://github.com/ao3575911/sinet.git
cd sinet
docker compose up -d --build
make build
./bin/SINet health
./bin/SINet
```

Read development OTPs in Mailpit at http://127.0.0.1:8025. Follow the keyboard dashboard
or use `SINet help` for JSON commands. Separate local sessions use SINET_CONFIG; the API
origin uses SINET_API. Credentials bind to the saved API origin.

```sh
./bin/SINet plug info echo 1.0.0
./bin/SINet plug install echo 1.0.0
./bin/SINet plug grant echo 1.0.0 repeat --approve
./bin/SINet agent call echo 1.0.0 repeat '{"text":"Hello SINet"}'
./bin/SINet agent audit
./bin/SINet plug revoke echo repeat
```

Sign in and claim a handle before the connector sequence. API default port is 8080.
All development ports bind to loopback. Development SMTP does not deliver to the Internet.
