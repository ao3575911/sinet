# Security invariants and threat model

## Assets and attackers

Protect email identities, session tokens, namespace/handle ownership, notes and direct/social content, model-provider secrets, tool grants, connectors, local host files, future wallets, and usage limits. Attackers include malicious registrants, connector publishers, compromised remote services, prompt-injection authors, and malicious agents/users.

## Non-negotiable controls

1. **OTP:** hashes stored, short TTL, one successful use, rate limit per IP + email + account, anti-enumeration responses. Production email delivery must be configured and tested.
2. **Handles:** DB-enforced unique normalized handle, transaction-safe claim, reserved names, Unicode homograph policy (ASCII-only initial), no reassignment without explicit workflow, auditing, no username squatting by unverified accounts.
3. **Tenancy:** every read/write filters by authenticated account/room membership; negative cross-tenant tests mandatory.
4. **Sessions:** revocable, short-lived or rotated, secure storage in OS keychain when possible (otherwise permission 0600, documented risks), no plain text credential logging.
5. **Connectors:** default deny; version-pinned manifest; per-operation grants; explicit approvals before mutation/data export; replay-safe confirmations; tool arguments validated; outputs size-limited; no unrestricted SSRF or arbitrary URLs. Remote HTTPS publishers must prove endpoint ownership; validate DNS + connected IP, deny private ranges/redirects and apply strict quotas.
6. **Prompt injection:** treat public profiles, room messages, third-party plugin descriptions, tool output, and web responses as untrusted; tool executor enforces authorization independent of model text.
7. **Local I/O:** cloud agent cannot read host filesystem or USB devices; explicit user-selected file/target required for future native bridge; symlinks and canonical paths checked; refuse surprising overwrite.
8. **Cloud terminal:** do not ship until sandbox is proven with isolation tests, per-session quotas, non-root execution, mount/egress policies, no Docker socket or host root access.
9. **Wallets:** future noncustodial design only; never ask the agent or connector to hold a recovery phrase; onchain payments require out-of-band confirmation.
10. **Safety:** reporting, blocking, abuse/spam handling, deletion/export, minimal logs, retention policy, content moderation path, spend controls.

## Public GitHub authorization

The reserved github-public contract permits only repository_summary with the fixed
owner/repository and external_disclosure risk. Gateway and GitHub are both disclosed.
The gateway, API adapter and CLI share strict public-fact validation; unknown fields,
invalid text/links/types and oversize responses cannot become successful tool output.
Per-account transaction locking serializes calls with grant/revoke/uninstall: already
authorized in-flight work can complete before revoke returns; fresh requests after
revocation commit cannot execute. Consent expiration uses clock_timestamp after lock
waits. Transport fixtures are injected only in tests; production entry points always
use DNS/connected-public-IP checked HTTPS with no proxies or redirects. No runtime
configuration enables private-network exceptions. See [evidence](releases/0.2.0-rc.1-validation.md).

## Security release checklist

The wallet/email expansion adds private, account-bound wallet ownership proofs and
email drafts. Wallet challenges bind the account, handle, origin, address, chain,
nonce and expiry; the authenticated API consumes them once and caps attempts.
Signatures, private keys and recovery phrases are not stored. The browser helper
uses a per-response CSP nonce, blocks application network connections and invokes
only personal_sign/account-selection methods. Wallet linking grants no transaction
or login authority. Gmail handoff requires explicit disclosure approval and leaves
review/sending in Gmail. See [the decision](ADR/0005-wallet-email-sections.md).

- Concurrent handle claims: exactly one winner.
- OTP replay, brute force, and expired codes fail.
- Unauthorized profile edits, room reads, notes reads, tool calls, install changes fail.
- Malicious connector manifest, unauthorized operation, mismatched version, private-network egress attempts fail.
- An instruction embedded in a message cannot override tool permissions.
- API tokens/secrets absent from logs and returned errors.
- Cloud agent has no host or USB capability in P0.
- Load and cost limits defined and verified before public signup.
