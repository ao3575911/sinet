# Acceptance and regression matrix

| Test | Expected result |
|---|---|
| Clean bootstrap | `docker compose up -d` + migrations + API + CLI work |
| New email OTP | Delivered through Mailpit in dev; verified once only |
| Expired/replayed OTP | Rejected |
| Two simultaneous claims | One handle winner; unique index prevents duplicate |
| Public profile | Shows intended fields, not private email or notes |
| Cross-user note read | Rejected |
| Room membership | Authorized join/read/post; abuse controls enforced |
| Connector registry | Valid manifest accepted for review; invalid rejected |
| Connector install | Only eligible reviewed manifest installed |
| Agent tool call | Missing grant denied regardless of model text |
| Revocation | Immediately denies subsequent calls |
| Connector external URLs | Private-network targets rejected for future HTTP adapters |
| No configured AI | Transparent unavailable state; no fabricated model output |
| Logs and errors | Do not expose credentials, OTP, session tokens |
| Quota exhaustion | Rejected gracefully; billed work not unbounded |
| CLI help | Human-readable and script-friendly |
| macOS/Linux/Windows compilation | Successful builds for three platforms |

## Quality checks

Go fmt/vet/test/race where supported; migrations applied from empty database; handler integration tests; replay and tenant isolation tests; static analysis for secrets; startup smoke test; reproducible build commands.
