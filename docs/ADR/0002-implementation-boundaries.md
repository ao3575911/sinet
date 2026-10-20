# ADR 0002: Persistent vertical slice and explicit approval

Status: accepted, 2026-10-20. Adopts ADR 0001 for this implementation.

Observed state: the root had a README and preserved reference library, no source code,
Go module, runtime or active contracts. Supplied sources remain unchanged. Active copies
of specifications live in docs/, contracts/, connectors/ and prompts/.

Use Go net/http, database/sql with lib/pq, PostgreSQL transactions and forward-only
embedded migrations. Use a JSON Schema validator for manifests and operation arguments.
No ORM, local daemon, general executor or uploaded-code interpreter.

ASSUMPTION: keep the literal `project-name` executable requested by the detailed contract;
SINet is the product name. Renaming and production domains await a release decision.

ASSUMPTION: a small keyboard menu plus line editor meets initial TUI needs; full-screen
rendering is unnecessary. Every dashboard action also has a noninteractive command.

Account writes serialize using transaction-scoped advisory locks. Handle uniqueness is
enforced by PostgreSQL. Idempotency records bind account, key, method, path and body;
reusing a key with different input fails. OTP redemption is intentionally one-use,
not replayed as a cached token response. Raw session tokens are never stored server-side.

Remote execution needs an operator-only CLI review action, an out-of-band hostname
ownership verification reference, and an explicit approved immutable version. The API
has no approval endpoint. Fixed POST operation routes append the operation name to the
manifest endpoint. HTTPS transport resolves and validates all IPs, pins the chosen public
IP, checks the actual peer, disables proxies/redirects, and bounds time and output.
Remote submissions always require fresh per-call consent, regardless of declared risk.
Consent is account/input/version bound, expires, and is consumed once. Model output
cannot issue executable tool requests; the initial model adapter is chat only. Users
invoke tools explicitly through the agent endpoint.

ASSUMPTION: one public `lobby`, 50 items per page, 40 model calls and 100 tool calls per
account per UTC day. Settings cannot relax consent. Email alias remains `unprovisioned`
without a controlled domain and routing; no fictional address or inbox is issued.

SMTP is a development Mailpit sink, or production SMTP requiring STARTTLS. A configured
OpenAI-compatible model is optional; missing configuration returns an explicit unavailable
state. Production delivery and a real model demo require external credentials. No claim
of public-launch readiness follows from prototype tests.
