# API v1 contract (draft)

Base path: `/v1`. HTTPS in non-local environments. JSON requests/responses. `Authorization: Bearer <token>` for protected calls. Use server-owned IDs, not user-supplied tenant IDs.

| Method | Route | Goal |
|---|---|---|
| POST | `/auth/request-otp` | Request verification code |
| POST | `/auth/verify-otp` | Redeem code, establish session |
| POST | `/auth/logout` | Revoke session |
| GET | `/me` | Own account + agent state |
| GET | `/sections`, `/sections/{id}` | Wallet/email availability and other section setup states |
| GET | `/wallets` | Own verified address links (maximum five) |
| POST | `/wallets/challenge` | Account/origin-bound five-minute ownership message |
| POST | `/wallets/verify` | Consume a personal_sign proof; does not issue a session |
| DELETE | `/wallets/{address}` | Remove own link with `{"confirm":true}` |
| GET | `/email` | Own verified sign-in email, Gmail handoff and alias status |
| GET/POST | `/email/drafts` | List/create private email drafts |
| GET/PATCH/DELETE | `/email/drafts/{id}` | Read/update/delete one owned draft |
| PATCH | `/me` | Edit public profile/settings |
| GET | `/handles/{name}/availability` | Advisory availability |
| POST | `/handles/claim` | Atomic verified claim |
| GET | `/users/{handle}` | Public profile |
| GET | `/users?q=` | Search discoverable users |
| GET | `/rooms` | List public rooms |
| POST | `/rooms/{slug}/join` | Join room |
| GET | `/rooms/{slug}/messages` | Paginated authorized messages |
| POST | `/rooms/{slug}/messages` | Send message |
| POST | `/reports` | Report content/identity |
| POST | `/blocks` | Block another account |
| GET/POST | `/notes` | Own notes |
| PATCH/DELETE | `/notes/{id}` | Own note update/delete |
| GET | `/connectors` | Browse public reviewed connectors |
| POST | `/connectors/submissions` | Submit manifest into review |
| GET | `/connectors/{id}` | Manifest and capabilities |
| POST | `/connectors/{id}/install` | Version-pinned installation |
| DELETE | `/connectors/{id}/install` | Uninstall and revoke grants |
| POST | `/connectors/{id}/grants` | Explicit allowed operation |
| DELETE | `/connectors/{id}/grants/{operation}` | Revoke operation |
| POST | `/agent/messages` | Agent message and audited tool calls |
| GET | `/audit/tool-calls` | Own recent tool activity |
| GET | `/healthz` | Unauthenticated minimal health status |

All list endpoints paginated; `409` conflict for handle collision; `401/403` for unauthenticated/forbidden; `429` with retry metadata for throttling. Idempotency keys required on sensitive write endpoints. Never return email in public profile.

## Implemented prototype extensions and details

The implementation follows the draft routes above. Its active command mapping is in
`SINet help`; operator-only actions are described in OPERATIONS.md.

- `GET /healthz` is the canonical health route; `/v1/healthz` redirects there.
- `GET /u/{handle}` is an escaped public HTML profile with OG title/description/URL.
- `GET /notes/{id}` also reads one owned note. `/notes` and other offset lists accept
  `limit` (1–50) and `offset` (0–100000). Directory ordering is handle; note ordering
  is update time/id. Messages/history/audit accept `before` (last numeric ID).
- `GET /connectors/installed`, `GET /connectors/submissions`, `GET /blocks`,
  `DELETE /blocks/{handle}`, `GET /agent/messages`, `GET /usage` are owner-scoped.
- `POST /agent/consents` accepts `{"tool":{...},"approve":true}`. Tool fields:
  `connector`, `version`, `operation`, `arguments`, optional `consent`.
- `POST /agent/messages` accepts either `{"tool":{...}}` or
  `{"message":"...","allow_model":true}`. Model text has no executable tool authority.
- Install requires `{"version":"1.0.0"}`; grant requires
  `{"version":"1.0.0","operation":"repeat","approve":true}`.
- `GET /me/export` returns bounded pages for each collection and
  `pagination.next_offset`. Continue until null; pause concurrent activity for a
  consistent offset-based export. `POST /me/delete` requires
  `{"confirm":"delete my account"}` and a valid idempotency key.
- `PATCH /me` accepts optional `display_name`, `bio`, `public`, `theme`, `notifications`.
  It does not accept account IDs, email changes, relaxed permissions or provider secrets.
- All authenticated mutations require `Idempotency-Key` (16–128 characters), including
  grant/revoke, logout and deletion. Store/retain keys in clients that retry uncertain
  writes. The shipped CLI generates fresh keys and never retries automatically.
- Lists return `{"items":[...]}`. Use the last ID for cursor lists; advance offset for
  offset lists. Usage returns the latest 50 daily counters; there is one lobby.
- Error codes are stable JSON `{"error":"code"}`. Invalid JSON/unknown fields fail.
  Empty success responses are not used. Public profiles exclude email and agent IDs.

Only built-in echo is preapproved. Remote service approval is an operator attestation
following out-of-band ownership and security checks, not automatic domain verification.
There is no API approval route, arbitrary execution endpoint, local file API or MCP runner.

## Unreleased wallet/email additions

See [wallet/email guide](WALLET_EMAIL.md). Challenge input is exactly `{"address":"0x..."}`;
response includes the canonical address, nonce, exact SIWE-format message, expiration and
fragment-based signing URL. Verify input is `{"nonce":"...","signature":"0x..."}`.
Successful proof links the address to the current account; it enables no wallet login,
transactions or balance access. Five attempts exhaust a challenge; a fresh one replaces
the old one. The static `GET /wallet/link` helper requires no bearer credential and cannot
submit a proof; challenge data must be passed in the fragment, not query parameters.

Draft create/update inputs are exactly `recipient`, `subject`, `body`. Recipient can be
empty. Gmail URLs are prepared by the CLI after explicit approval; there is no send or
inbox-reading API. Private export includes wallets/drafts; public profiles omit them.

## Public GitHub operation (0.2.0-rc.1)

`github-public@1.0.0` reserves exactly `repository_summary` with external_disclosure
risk and `contracts/github-public/parameters.schema.json`. No live manifest is bundled.
Consent responses retain the destination host field and add connector/version, exact
arguments, disclosure and recipients (the reviewed gateway URL and https://api.github.com).
Successful output is revalidated before audit success; fixed `github_*` failures carry
no remote error body. Calls use the existing private agent and audit endpoints.
See [user guide](GITHUB_CONNECTOR.md) for revocation/queued-expiry semantics and live gates.

## Security review hardening (2026-10-22)

Requests reject duplicate JSON object keys, invalid UTF-8, trailing values, nesting
beyond 64 levels, number literals longer than 128 bytes and exponents outside
-308..308. Generic connector consent preserves exact number literals, including
integers above 2^53; key order and whitespace remain insignificant. Changing a
number's representation may require fresh consent, even for mathematically equal
values. CLI requests and displayed API results preserve those number literals too.

Session validity is rechecked against wall-clock time after the account lock, before
idempotency replay or execution. A request whose session expires in that queue gets
401 and cannot write. Session/rate-counter database failures return 503, not false
401/429 responses. Public profile database failures also return 503 rather than 404.
Health and public HTML profile database calls have explicit 2s/5s deadlines.

Parameter schemas must be inline: 8192 bytes, 512 value nodes, depth 16 maximum;
`$ref`, `$recursiveRef` and `$dynamicRef` keys are rejected anywhere in the parameter
document. This conservative bound also applies to schema metadata/constant data.
Existing manifests requiring references must be rewritten and reviewed as a new
immutable version; they fail closed until then. Bundled echo and github-public
contracts already meet these limits. Competing publishers cannot claim separate
versions of the same connector ID concurrently.

See [the review and regression evidence](SECURITY_REVIEW_2026-10-22.md).
