# API v1 contract (draft)

Base path: `/v1`. HTTPS in non-local environments. JSON requests/responses. `Authorization: Bearer <token>` for protected calls. Use server-owned IDs, not user-supplied tenant IDs.

| Method | Route | Goal |
|---|---|---|
| POST | `/auth/request-otp` | Request verification code |
| POST | `/auth/verify-otp` | Redeem code, establish session |
| POST | `/auth/logout` | Revoke session |
| GET | `/me` | Own account + agent state |
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
