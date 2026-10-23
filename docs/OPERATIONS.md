# Running and operating the prototype

## Local stack

Use a current patched Go compiler (validation target: **Go 1.27.3**) and Docker Compose.
No cloud account is needed for the deterministic MVP. Do not use the development defaults
as deployment credentials.

```sh
docker compose up -d --build
make build
./bin/SINet health
./bin/SINet
```

Compose starts PostgreSQL, Mailpit and a nonroot API with a read-only root filesystem,
no capabilities, no privilege escalation and no host-data or Docker-socket mounts.
The API applies checksummed SQL migrations under a database advisory lock at startup.
Database state uses a named volume. CLI binaries have no mandatory local daemon.

API: http://127.0.0.1:8080. Mailpit: http://127.0.0.1:8025. Mailpit is a development SMTP
sink; email is not sent to the Internet. All published ports bind to loopback.
The Compose file embeds local-only credentials. Do not expose its database or Mailpit.

For a managed environment with an intercepting TLS proxy, retain the runtime proxy and
mount its public CA only during dependency download:

```sh
docker build --secret id=proxy_ca,src="$CODEX_PROXY_CERT" -t sinet-api:prototype .
docker compose up -d --no-build
```

The CA is not copied into image layers. A network denial is a configuration dependency,
not permission to disable TLS or change routes. For native API development instead:

```sh
docker compose up -d db mailpit
make build
./bin/server migrate
./bin/server
```

The server reads environment variables; it does not automatically source `.env`.
`.env.example` lists configuration. `APP_ENV=production` requires HTTPS public URL,
PostgreSQL `sslmode=verify-full` as one parsed URL parameter (duplicates and query-host overrides rejected), SMTP settings with mandatory STARTTLS, and a random
`OTP_HASH_KEY` of at least 32 characters. The public URL must be a bare HTTPS origin without credentials, query or fragment. Development uses an ephemeral OTP hash key;
restarting the API invalidates pending development OTPs, while sessions remain valid.
Place a trusted TLS reverse proxy in front of the API before any remote use. The API
ignores forwarded IP headers; behind a proxy, IP throttles apply to the proxy address.
Configure edge throttles deliberately before public signup.

## First connector outcome

Interactive: run `SINet`, verify email using Mailpit, claim a handle, browse the
connector menu, inspect/install `echo` version `1.0.0`, grant `repeat`, and run it.

Equivalent scriptable commands:

```sh
./bin/SINet login request you@example.test
# View the code in Mailpit. Feed the code on stdin; it is not a command-line argument.
./bin/SINet login verify you@example.test
./bin/SINet claim your_handle
./bin/SINet room lobby join
./bin/SINet room lobby say 'Hello SINet'
./bin/SINet note new 'First note' 'Saved in PostgreSQL'
./bin/SINet plug list
./bin/SINet plug info echo 1.0.0
./bin/SINet plug install echo 1.0.0
./bin/SINet plug grant echo 1.0.0 repeat --approve
./bin/SINet agent call echo 1.0.0 repeat '{"text":"Hello SINet"}'
./bin/SINet agent audit
./bin/SINet plug revoke echo repeat
./bin/SINet share
```

Echo is a deterministic tool, not AI. Calls before installation/grant and after revocation
fail. Chat without a provider explicitly returns `mode: no_model`. Set all three of
`MODEL_API_BASE_URL` (including `/v1` where required), `MODEL_API_KEY`, `MODEL_NAME` to enable
OpenAI-compatible chat. `agent ask 'text' --allow-model` authorizes sending the conversation
to that operator-configured provider; last 20 messages may accompany it. User settings
cannot remove grant or consent checks. Model output never automatically runs tools.
No real provider credentials have been used for validation.

`SINET_CONFIG` selects a separate account/session file; `SINET_API` selects
the server. Session files contain bearer credentials, bind to one API origin, use atomic
replacement and mode 0600 on Unix, and reject symlinks/insecure permissions. An OS keychain
is not integrated. On Windows, verify user-only filesystem ACLs; Unix mode checks are not
portable to Windows. Credentials are not encrypted at rest. Logout revokes this session;
account deletion revokes all sessions. Normal token expiry is 24 hours, checked again
after account-lock waits. Session files must be regular files of at most 4096 bytes
with a 64-character hex token. Use a private, trusted config directory; these checks
do not protect against a compromised local account or hostile parent directories.

## Connector review and remote services

`plug validate FILE` and `plug submit FILE` are explicit user-directed local file reads.
No server endpoint accepts executable uploads. Submissions validate against the supplied
JSON Schema and compiled operation schemas, start in `review`, and cannot overwrite a
published version. Parameter schemas must be inline and meet the [documented resource limits](API.md#security-review-hardening-2026-10-22). Publisher must match the authenticated handle; concurrent submissions serialize ownership checks for each connector ID. Only reviewed versions
can be installed; changing versions requires removal/reinstallation and new grants.

Operators have database privileges and must run review commands in a trusted environment:

```sh
./bin/server review list
./bin/server review approve CONNECTOR_ID VERSION 'ownership: verification-record; review: security-review-record' --confirm
./bin/server review reject CONNECTOR_ID VERSION 'incident/revocation-reference' --confirm
```

Before approval, independently verify publisher control of the endpoint hostname via a
reviewed DNS TXT or HTTPS ownership challenge tied to the submitted account, connector and
version. Record the exact verification and security review evidence. The command records
an operator attestation; it does **not** automatically prove hostname ownership or vet a
service. No real remote service is approved by this repository. Submitted builtins cannot
be approved; supported builtin code ships with the server.

For reviewed HTTPS services, fixed POST routes append `/OPERATION` to `endpoint_url`.
Arguments and outputs must be JSON. HTTPS port 443 only, no URL credentials/query/fragment,
no proxy/redirect, public DNS answers only, pinned IP with peer recheck, TLS 1.2+, 8-second
request timeout, 16 KiB headers, 8 KiB arguments and 64 KiB output budget. Private,
link-local, loopback, multicast, reserved and common embedded-IP transition ranges are
blocked. Network infrastructure egress rules are still required for a public deployment.
No production connector authentication/OAuth secrets or retries are implemented.

Every remote call needs a fresh 5-minute consent token bound to account, version, operation
and exact arguments, even when its publisher labels it read-only. The CLI previews
unapproved calls and requires `--approve` to create/consume that consent. Writes and exports
also require per-call consent. Successful calls and denials have owner-only audit records.
Revocation prevents subsequent calls; a call already dispatched cannot undo remote effects.
Because external side effects and PostgreSQL commits are not a distributed transaction,
a remote success followed by process/commit failure can leave outcome uncertain. Do not
retry a privileged external operation automatically. Require upstream idempotency and
reconciliation before approving services that perform valuable irreversible actions.

## Abuse, quotas and retention

One public `lobby` requires a claimed handle and membership. Messages are paginated, not
live streamed. Blocking filters directory and room visibility in both directions. It does
not hide deliberately public profile pages from anonymous visitors. There are no direct
messages, automated moderation or push notifications. The notification setting is a stored
preference only. Privacy setting removes the profile from public JSON/HTML/directory.

Reports enter an operator queue. Operators can resolve a report, optionally deleting its
message. There is no automatic public account-suspension system.

```sh
./bin/server reports list
./bin/server reports resolve REPORT_UUID keep --confirm
./bin/server reports resolve REPORT_UUID remove-message --confirm
./bin/server metrics
./bin/server maintenance --confirm
```

Schedule maintenance daily through a trusted scheduler. It is deliberately not an automatic
background deletion process. The command prunes expired OTPs (after one day), sessions,
consents, idempotency and throttle rows; chat after 30 days, room/tool/usage records after
90 days, reports and identity audit after 365 days. Chat also keeps at most 200 messages
per account. Notes/settings persist until deletion. Deleted handles remain reserved against
impersonation, with no account link. Funnel metrics contain only daily aggregate counts,
not identities, emails, note bodies, chat prompts or tool arguments.

Enforced per-account limits per UTC day: 40 model calls, 100 tool attempts, 200 room posts,
20 reports, 10 manifest submissions. At most 100 notes and 50 installed connectors. Request
bodies are limited to 128 KiB, notes to 64 KiB, messages to 4000
characters; list pages contain at most 50 items. Mutating account requests are throttled
at 120/minute. OTP delivery: 10/IP/hour, 3/email/15 minutes; verification: 30/IP/15 minutes,
15/email/15 minutes and 5 attempts/code. SMTP delivery failures are explicit. OTPs last five
minutes and are one-use under concurrent redemption. SMTP and model secrets never enter
public responses. Model output cap is 512 tokens; quotas count attempts, including upstream
failures. Monetary billing limits still need provider-side setup.

Mutating authenticated API calls require a 16–128 character `Idempotency-Key`. Reuse with
different method/path/body conflicts; matching successful writes replay for 24 hours.
The CLI generates one key per invocation and does not automatically retry. Use API clients
with a retained key for uncertain sensitive writes. OTP redemption is one-use and does not
cache a raw session token. Generic successful API response caching includes private app
data and consent tokens; protect the database and its backups accordingly.

## Backup, export, deletion and recovery

`SINet export` returns account data and pages of notes, messages, grants, chat and tool audit.
Repeat with `export --offset NEXT_OFFSET` while `pagination.next_offset` is non-null;
concurrent writes can shift offset pages, so pause activity while exporting. Each table
uses the same page size (default 20, maximum 50).
protect the exported JSON as private. `delete-account --confirm-delete` deletes owned live
rows, sessions and private content. Shared report references are detached; published
connector metadata is anonymized and disabled. Aggregate counters and detached audit events
are retained. Handle tombstones remain reserved. Deletion from existing backups follows
backup expiry, not immediate cryptographic erasure.

Example trusted operator backup/restore (backup files contain private data):

```sh
mkdir -p work/backups
umask 077
docker compose exec -T db pg_dump -U sinet -Fc sinet > work/backups/sinet.dump
# Restore into a NEW database first; never overwrite the live database to test recovery.
docker compose exec -T db createdb -U sinet sinet_restore
docker compose exec -T db pg_restore -U sinet -d sinet_restore < work/backups/sinet.dump
```

Keep encrypted off-host backups with access controls and a documented expiry. Test recovery
regularly. Migrations are forward-only and checksum-protected; changed historical migrations
fail startup. Database migrations and backups require operator access. No deployment,
controlled mail domain, routed alias/inbox, OAuth integration, VM sandbox, payments, wallet,
USB bridge, arbitrary shell or host filesystem agent capability is shipped.

## Signing

For the current release-key custody, rotation, revocation and recovery procedure,
see [native validation and signing lifecycle](NATIVE_CREDENTIALS.md#dedicated-release-signing-key-lifecycle).
The following records the original prototype signing setup.

All implementation commits use SSH signatures. Public fingerprint:
`SHA256:DZbp/eHiq7txsnGklKuBtZ1XOotv927BpKkCIfS5ggw`.
The owner registered this key as a GitHub signing key. GitHub API verification and exact
remote SHA checks are recorded in IMPLEMENTATION.md. GitHub does not support Git signed
push certificates; commits are signed and transported over authenticated HTTPS.
The dedicated private signing key is only in ignored `work/signing/`, with mode 0600;
it is not in source, Docker context, binaries or downloadable deliverables. Securely
transfer it outside the workspace if future signing with this identity is needed, or
remove its registration after this task. Do not disclose the private key in chat.


## Existing database volumes after the SINet rename

Fresh installations use database/role `sinet`. PostgreSQL does not rename an existing
volume's role or database when environment variables change. For a volume created before
this rename, add these non-secret compatibility settings to the ignored local `.env`:

```dotenv
POSTGRES_USER=projectname
POSTGRES_DB=projectname
```

Compose uses these settings for both readiness and the API connection. When running the
API or integration tests directly, set DATABASE_URL / TEST_DATABASE_URL to the existing
role/database as well. Substitute those identifiers in backup commands above. No database
or role is renamed or deleted automatically. Retain backups before an operator-led rename.
The old echo publisher in migration 0003 is historical; migration 0004 updates it to SINet.
