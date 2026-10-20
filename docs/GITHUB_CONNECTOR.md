# Public GitHub gateway: implementation slice

Status: implemented locally for review, not deployed or approved in the connector
registry. Target application release remains 0.2.0. This advances the existing
[next-slice plan](NEXT_SLICE.md), steps 2–3, without changing grants, consent,
manifests in the live registry, migrations or the published CLI release.

## Contract and disclosure

Proposed connector: `github-public@1.0.0`; operation: `repository_summary`;
risk: `external_disclosure`. The gateway implements `POST /repository_summary`,
compatible with the existing fixed-endpoint connector adapter. Request JSON is
exactly `{"owner":"ao3575911","repository":"sinet"}`. Unknown/duplicate fields,
other repositories, trailing JSON and bodies above 512 bytes fail before upstream
I/O. The closed schemas are in `contracts/github-public/`.

Proposed user disclosure, to be incorporated into a reviewed manifest after
hostname provisioning:

> Send owner "ao3575911" and repository "sinet" to the approved SINet GitHub
> gateway and GitHub to fetch public repository facts. No SINet session, email,
> notes, profile or model conversation is sent. This operation performs no writes.

`GET https://api.github.com/repos/ao3575911/sinet` is the only upstream request.
It is constructed independently of inbound headers and accepts no GitHub token,
host, URL, account data or credentials. Upstream requests contain only Accept,
API-version and User-Agent application headers, plus the transport's standard
HTTP headers. Caller Authorization/Cookie/forwarded-IP headers are not copied.

The output includes canonical repository name/link, nullable description and
language, stars/forks/open-issue counts, archived state, upstream timestamp and
source URL. GitHub's open-issue count includes pull requests. Only explicitly
selected fields are returned; unknown upstream fields are discarded. Private or
different repositories, missing/wrong-type fields, negative or oversized counts,
invalid timestamps and terminal/control-format characters fail closed. Text is
bounded in UTF-8 bytes. Prompt-like descriptions remain inert untrusted data;
there is no model invocation, operation chaining or permission grant in the gateway.

## Resource limits and failures

The shared connector HTTPS client pins public DNS results to the connected IP,
denies private/mixed DNS and peer mismatches, and disables redirects and proxies.
It enforces TLS 1.2+, an eight-second overall timeout, header/TLS/dial budgets and
16 KiB response headers. Decoded upstream response bodies are capped at 64 KiB.
No automatic retries, cached-success substitution or upstream error-body forwarding
are used. At most four calls are in flight and 60 calls start per minute per
gateway process. These are service-wide budgets, not per-account authorization.

| Failure | HTTP response |
| --- | --- |
| Invalid arguments / oversized input / wrong media type | 400 / 413 / 415 |
| Repository not found | 404 `repository_not_found` |
| GitHub 403, or 403/429 rate limit | 503 `upstream_forbidden` / `upstream_rate_limited` |
| Timeout | 504 `upstream_timeout` |
| Other upstream/network/validation failure | 502 with a fixed error code |
| Four in-flight requests / local minute budget exhausted | 503 `gateway_busy` / 429 `gateway_rate_limited` |

Responses are JSON with `no-store` and `nosniff`. The standalone server logs
startup/fatal server errors, not request bodies or caller headers. Inbound read,
header, write, idle and shutdown limits are configured. `GET /healthz` reports
process readiness only and makes no GitHub call.

## Run and validate

```sh
go test -race ./internal/githubconnector ./internal/connectors
go build -trimpath -o bin/connector-github ./cmd/connector-github
./bin/connector-github # binds only 127.0.0.1:8090 by default
curl -fsS http://127.0.0.1:8090/healthz
curl -sS -H 'Content-Type: application/json' \
  -d '{"owner":"ao3575911","repository":"sinet"}' \
  http://127.0.0.1:8090/repository_summary
```

The final command requires an operator network that permits the fixed GitHub
HTTPS destination. The gateway deliberately does not use environment proxies;
it must not be used to bypass the managed environment's outbound policy. Fixture
tests inject an in-memory transport and do not establish live-service evidence.
Fixtures are explicitly synthetic, dated 20 October 2026. Tests cover successful
projection, invalid input with zero upstream calls, nullable fields, hostile text,
malformed/private/oversized data, error mapping, header isolation, redirects,
timeouts, cancellation, concurrency and rate budgets. Existing connector tests
cover the shared DNS and peer-IP checks.

## Remaining approval and delivery gates

Provision a controlled public hostname with HTTPS on 443 and record ownership/TLS
evidence and a public privacy policy. Configure egress allowing only GitHub, an
ingress service identity or equivalent abuse protection, monitoring, quotas and
rollback. Run non-root with a read-only filesystem, dropped capabilities and no
host or Docker socket mounts. `SINET_GITHUB_LISTEN` controls the bind address
behind that reviewed TLS ingress; changing it does not enable registry approval.

Only then create and review the immutable manifest and test the complete existing
install → grant → input/version/account-bound one-use consent → agent call → audit
→ revoke → deny journey using two accounts and counted upstream calls. The
gateway is not an authorization replacement. No localhost exemption or automatic
manifest approval is added. No live outcome or production readiness is claimed.

## Validation record — 20 October 2026

Validated locally with the official Go 1.27.3 Linux amd64 toolchain (download
SHA-256 checked), PostgreSQL 17.6 and Mailpit 1.27.10:

- `make check integration cross`: formatting, vet, module verification, full
  PostgreSQL race tests and all three existing CLI cross-builds passed.
- Gateway schema/security/budget/cancellation tests passed under the race detector.
- Seven Python publication/version tests, two Node automation tests and actionlint
  1.7.12 passed.
- Gateway process health returned 200 and a forbidden repository returned 400.
- Real two-account CLI/SMTP/PostgreSQL and POSIX PTY journeys passed. Four database
  migrations were applied; zero smoke-test accounts remained after cleanup.
- govulncheck 1.9.0 could not complete locally: the managed HTTP policy returned
  `Forbidden` for `https://vuln.go.dev/index/modules.json.gz`. CI's vulnerability
  result must be checked separately; this is not a clean-scan claim.

No production hostname, live GitHub gateway journey, native desktop UX validation
or new release is established by these checks. The npm 0.1.1 package remains
unchanged. Existing PyPI 0.1.1 Linux wheel bytes were independently downloaded,
checked against the public registry SHA-256 and installed in an isolated venv;
its `SINet version` command returned 0.1.1.
