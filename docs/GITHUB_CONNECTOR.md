# Public GitHub gateway: implementation slice

Status: complete local CLI/TUI/API integration in **0.2.0-rc.1** source, not
published, deployed or approved in a live registry. No model credentials are needed.
See [sprint checklist](GITHUB_SPRINT.md), [validation evidence](releases/0.2.0-rc.1-validation.md)
and [deployment/operator review](GITHUB_DEPLOYMENT.md). The existing wallet ownership
and Gmail draft features are preserved. npm/PyPI packages remain 0.1.1.

## User journey after operator approval

Build current source; the published 0.1.1 installers do not contain this journey.
Sign in and claim a handle, then run:

```sh
SINet plug info github-public 1.0.0
SINet plug install github-public 1.0.0
SINet plug grant github-public 1.0.0 repository_summary --approve
SINet agent github
# Read both recipients, exact owner/repository, purpose and privacy disclosure.
# The preview exits with approval_required and makes no upstream call.
SINet agent github --approve
SINet agent audit
SINet plug revoke github-public repository_summary
SINet agent github --approve # denied; no upstream request
SINet plug remove github-public
```

The explicit equivalent is `agent call github-public 1.0.0 repository_summary
'{"owner":"ao3575911","repository":"sinet"}' [--approve]`. A different repository,
extra/duplicate input fields or another operation fails. Each approved command obtains
one fresh five-minute consent for this account, pinned connector version, operation and
exact input. Failed upstream attempts consume consent too; no automatic retries occur.

In the dashboard choose **5. Connectors → 8. GitHub public facts and audit**. Inspect,
install, grant, invoke, audit, revoke and uninstall are separate steps. Invoking shows
both recipients and exact fields, then requires typing `invoke`. Cancelling makes no
upstream request. An unavailable reviewed version tells you an operator must deploy and
approve it. The fixture gateway is never registered in your development database.

Results are bounded JSON facts with a source URL and upstream updated timestamp. They
are labelled untrusted data; descriptions cannot grant permissions or trigger tools.
The API validates the returned gateway contract before recording success; the CLI also
validates it before rendering. Unknown/duplicate fields, wrong links/types, invalid text
and oversize data fail closed. No tool results enter model history.

Revocation/uninstall use the existing per-account transaction lock. An invocation that
already holds it may finish before revoke returns; after revoke commits, fresh calls
are denied. Re-granting does not resurrect old consents. Consent expiry is checked with
wall-clock time after any lock wait. Reusing an idempotency key can return a previous
committed response without executing another request; it does not restore permission.

## Contract and disclosure

Proposed connector: `github-public@1.0.0`; operation: `repository_summary`;
risk: `external_disclosure`. The gateway implements `POST /repository_summary`,
compatible with the existing fixed-endpoint connector adapter. Request JSON is
exactly `{"owner":"ao3575911","repository":"sinet"}`. Unknown/duplicate fields,
other repositories, trailing JSON and bodies above 512 bytes fail before upstream
I/O. The closed schemas are in `contracts/github-public/`.

The CLI preview, TUI confirmation and server consent response disclose both recipients;
the actual gateway endpoint comes from the immutable reviewed manifest:

> Send owner "ao3575911" and repository "sinet" to the approved SINet GitHub
> gateway and GitHub to fetch public repository facts. No SINet session, email,
> notes, profile, model history or provider secrets are sent. This operation performs no writes.

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

## Remaining live gate

Provision a controlled public hostname and hosting access, verify ownership/public TLS
and operational controls, freeze the real immutable manifest, and obtain a separate
operator approval. Then run the real two-account journey against that approved HTTPS
gateway and GitHub, recording redacted audit evidence. No hostname was supplied, so
no placeholder manifest was created or approved. Local integration is proven with
isolated synthetic fixtures; no live GitHub journey or production readiness is claimed.
See [deployment instructions](GITHUB_DEPLOYMENT.md).

The API returns locally selected `github_*` codes for not-found, forbidden, rate limits,
busy service, timeout and invalid upstream responses. Other network/egress failures use
a fixed unavailable/destination-denied code. Remote error bodies are never forwarded.
Failures are audited as authorized execution_failed, rather than fabricated success.

## Validation record — 22 October 2026

[Release-candidate ledger](releases/0.2.0-rc.1-validation.md) records real PostgreSQL
race tests, counted two-account adapter/gateway fixtures, CLI/PTY success and cancellation,
SMTP and wallet browser regressions, migration state and vulnerability results. The
compiler is now Go 1.27.3; the initial 1.27.2 scan found a reachable x509 advisory and
was fixed. No reachable vulnerabilities remain in the patched scan; a module-only
OpenPGP advisory remains in an unused package. Earlier results below are historical.

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
