# Public GitHub gateway deployment and review

Prepared configuration, not a deployed service. Chosen design: one dedicated Linux
container host with an operator-managed HTTPS ingress on TCP 443. No cloud account,
controlled hostname, ownership proof or operator attestation is available in this sprint.
Do not submit/approve a manifest until those dependencies exist. Deployment and SINet
registry approval are separate decisions. This service needs no GitHub/model credentials.

## Build and operate

Build from the reviewed signed sprint commit using the patched Go 1.27.3 Dockerfile:

```sh
docker build -f deploy/github-public/Dockerfile -t sinet-github:review .
```

Record the source SHA, base-image digests, final image digest, build compiler and image
scan result. Resolve mutable base tags to reviewed digests for the actual deployment;
store the built image in the operator's controlled registry. Select the resulting
`registry/repository@sha256:...` reference as SINET_GITHUB_IMAGE. No image was published
by this sprint. Use the same digest when restarting or rolling back; never roll a tag.

```sh
docker compose -f deploy/github-public/compose.yaml config --quiet
docker compose -f deploy/github-public/compose.yaml up -d --no-build
curl -fsS http://127.0.0.1:18090/healthz
docker compose -f deploy/github-public/compose.yaml ps
```

The gateway listens only on host loopback port 18090. Host ingress forwards only
GET /healthz and POST /repository_summary from the owned HTTPS hostname to that
port. Terminate TLS 1.2+ with a valid public certificate; keep HTTP invocation disabled.
Configure 16 KiB header and 512-byte request budgets, a 15-second upstream deadline,
edge connection/rate limits and an ingress allowlist for SINet API egress addresses
or equivalent reviewed service identity. Do not add an authorization header that would
require forwarding a user's SINet session. Strip Authorization/Cookie/forwarded identity
headers at ingress; the gateway independently constructs its GitHub request.

The compose service uses UID/GID 65532, a read-only filesystem, no volumes or host/Docker
socket mounts, all capabilities dropped, no-new-privileges, 128 MiB memory, 0.5 CPU,
32 PIDs, 256 descriptors and bounded container logs. GOMEMLIMIT is 96 MiB. Four requests
may run concurrently and at most 60 start per minute per process. Keep one replica until
an operator reviews shared abuse budgets; replicas multiply the gateway limit.

Permit DNS through an approved resolver and outbound HTTPS only to api.github.com on
443 using platform domain-aware egress policy. Public-DNS/connected-IP validation, TLS,
no proxies, no redirects and response/time budgets remain enforced by the Go transport.
Do not introduce private-network exceptions, proxies or alternate IP routes to bypass a
managed policy. Restrict the host firewall as well as container ingress; the container
configuration alone does not enforce destination egress or provide public TLS.

## Health and monitoring

`/connector-github healthcheck` probes only local /healthz with a two-second deadline;
it makes no GitHub request. The container checks every 30 seconds, fails after three
failures and allows ten seconds for shutdown. Process readiness is distinct from GitHub
availability. Docker does not automatically restart an unhealthy running process: the
operator must alert and investigate, then explicitly restart or roll back.

Before approval, configure alerts for readiness failures, restarts/OOM, sustained CPU or
memory budget pressure, ingress 429/502/503/504 rates and latency above eight seconds.
Record fixed status/error codes and aggregate counts only; disable request/response body,
query, Authorization and Cookie logging. Bound log retention and access. An explicit
scheduled synthetic fetch of the one fixed repository can check upstream availability;
include it in the 60/min budget and label its audit/service-account evidence. Health
probes alone cannot establish live upstream readiness or two-account authorization.

## Immutable manifest and operator review

When a controlled endpoint is known, create one manifest with:

- ID github-public, version 1.0.0, runtime remote_https; publisher is the actual submitting
  account's claimed handle, with a public privacy policy describing both gateway and GitHub.
- Actual owned HTTPS endpoint root (without /repository_summary), no query/userinfo or
  custom port; the adapter appends the fixed operation.
- Exactly one operation repository_summary, risk external_disclosure, and the exact
  `contracts/github-public/parameters.schema.json` object. Output follows its companion
  output schema and the stricter UTF-8 byte/control-character checks.
- Operation description (max 280 characters): "Send owner ao3575911 and repository sinet
  to this gateway and GitHub for read-only public repository facts; no account data."

No placeholder manifest is generated or approved here. Freeze the actual JSON and
record `sha256sum manifest.json`. Validate and submit through `SINet plug validate`
and `SINet plug submit`. Submission stays in review. Published versions are immutable;
changing endpoint, schema or recipients requires a new reviewed version, reinstall and
a new grant/consent. Do not overwrite an existing version.

The operator independently verifies account/version-bound DNS TXT or HTTPS ownership,
public DNS addresses, the actual connected peer, certificate chain/hostname/expiry,
redirect denial, fixed route, privacy policy, ingress/egress, resource controls and
monitoring. Store timestamped evidence and a substantive review reference. Only then:

```sh
./bin/server review list
./bin/server review approve github-public 1.0.0 'ownership:EVIDENCE_REFERENCE review:REVIEW_REFERENCE' --confirm
```

The command records an attestation; it cannot prove hostname ownership itself. Never
use the synthetic references from automated tests for an actual registry approval.
See [user journey](GITHUB_CONNECTOR.md) and [acceptance gates](NEXT_SLICE.md).

## Live evidence and rollback

Run the two-account install/grant/disclose/consent/agent/audit/revoke/uninstall journey
against the approved endpoint and real GitHub. Capture UTC timestamps, source/image and
manifest digests, redacted commands, returned facts/source/time, per-account audit IDs,
allowed/denied counts and ingress observations. Record local fixtures and live observations
in separate ledgers. Run adversarial and counted upstream cases in isolated fixtures;
never treat synthetic GitHub payloads as live evidence. No wallet transaction/email send.

Before any update, save the previous reviewed image digest/config and manifest/review
references. If unsafe behavior appears, reject the connector version first:

```sh
./bin/server review reject github-public 1.0.0 'review:INCIDENT_REFERENCE rollback:CHANGE_REFERENCE' --confirm
```

Calls check approved state server-side. Calls already authorized and in flight can finish;
rejection cannot retract a request already sent. Quiesce ingress before service changes.
Restore SINET_GITHUB_IMAGE to the previous recorded digest and run compose up -d
--no-build, verify local and public HTTPS health, inspect error rates and complete a fresh
consented smoke invocation before resuming. Rejected versions cannot simply be approved
again: submit a new immutable version and obtain fresh installs, grants and consent.
The gateway has no persistent application state or migrations. Back up SINet PostgreSQL
using existing operations guidance; this sprint adds no database migration.
