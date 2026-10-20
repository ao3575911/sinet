# Wallet and email sections

Unreleased source feature targeting 0.2.0. The published npm/PyPI 0.1.1 CLI does
not contain these commands yet. Build current source to try them. In the
dashboard, choose **8. Wallet, email and services**. Every action also has a plain
CLI command. [Design and security decision](ADR/0005-wallet-email-sections.md).

## Link your existing wallet

```sh
SINet wallet connect 0xYOUR_ETHEREUM_ADDRESS
SINet wallet list
SINet wallet unlink 0xYOUR_ETHEREUM_ADDRESS --confirm-unlink
```

Sign in with email and claim a handle first. `connect` prints a one-time message
and browser signing link. Open it on the configured SINet API site in a browser
with your Ethereum wallet extension. Select the displayed address and Ethereum
Mainnet, review the SINet handle and message, and sign. Copy the returned signature
into the CLI. Your wallet keeps its private keys and recovery phrase.

Scriptable flow:

```sh
SINet wallet challenge 0xYOUR_ETHEREUM_ADDRESS
# Sign the exact returned message with personal_sign in your wallet.
# Paste the signature through stdin, using the returned nonce:
SINet wallet verify NONCE
```

This proves control of an Ethereum EOA; it does not submit a transaction, show a
balance, authorize purchases, issue a session or expose your address publicly.
Contract wallets and other chains are not supported in this first slice. Up to
five addresses can be linked. A challenge expires after five minutes, is bound
to your account and site, and can succeed once. Requesting another replaces it.
Unlinking removes the identity link; funds remain with the external wallet.

The signing page cannot submit proofs or access SINet sessions. Its only wallet
methods are account selection, chain selection checking and personal_sign. It
uses a strict per-response script nonce, blocks application network connections
and framing, removes the URL fragment and renders content as text. The server
checks the proof independently, including expiry, origin, nonce and account.

## Email with Gmail

```sh
SINet email status
SINet email new friend@example.test "Hello" "Draft text"
SINet email list
SINet email get DRAFT_ID
SINet email edit DRAFT_ID friend@example.test "Hello again" --stdin
SINet email gmail DRAFT_ID --approve
SINet email inbox
SINet email delete DRAFT_ID
```

The composer saves private drafts. Opening the prepared Gmail link shares the
recipient, subject and body with Google. Review your sender account in Gmail,
then send there. SINet continues to label the local record as a draft: it cannot
observe whether you sent it. The inbox command returns your external Gmail inbox
link; Gmail remains responsible for login, sending and receiving. Links are
prepared locally and are never opened or sent automatically.

Blank recipients are allowed while drafting. Subjects are limited to 200 UTF-8
bytes and bodies to 65536 bytes; at most 100 drafts are saved. Gmail compose links
have an 8192-byte encoded URL limit. Draft headers reject line/control injection;
bodies accept newlines/tabs and normalize CRLF. Lists accept --limit/--offset.
Wallet links and drafts are included in private account exports and deleted with
the account. They are excluded from public profiles.

Automatic inbox sync/sending needs Google OAuth setup and per-user authorization.
No Gmail password, provider token or wallet secret is requested by this flow.
The verified email shown in status is your SINet OTP sign-in address. A separate
SINet email alias stays unprovisioned until an owned domain and routing exist;
this feature cannot create addresses on Google's gmail.com domain.

## Other sections

`SINet sections` or `SINet sections ID` shows each section and its next setup step.
Business, ENS/external identity, files/devices, terminal/cloud and mobile remain
planned. Existing notes, profiles, social/connector tools and settings continue
to work. Wallet proof and email drafts do not grant host access, cloud compute,
identity resolution, payments or business-domain mail capabilities.

## Validation

The database-backed race suite covers account isolation, origin/nonce/message
binding, expiry/replacement, bounded attempts, concurrent proof consumption,
duplicate identity prevention, unlink, draft pagination/limits/idempotency,
export and deletion. Unit tests use an independent public ethers personal_sign
fixture and EIP-55 vectors, including high-s and malformed-signature denials.
Gmail tests cover fixed-origin URL encoding and explicit approval.

The PTY acceptance script creates a draft and prepares a Gmail link after
disclosure. `tests/wallet_browser.py` uses Playwright 1.62.0 and Chromium with a
clearly mocked EIP-1193 provider to check exact signing bytes, account/chain/purpose
checks, fragment removal and no external requests. It does not prove a live wallet
extension or Gmail delivery. Run it with a local API/Mailpit and built CLI:

```sh
python3 -m pip install playwright==1.62.0
python3 -m playwright install --with-deps chromium
python3 tests/wallet_browser.py
```

Live wallet-extension validation, Google OAuth approval/synchronization and domain
provisioning remain separate gates. No mainnet transaction or real email was sent
by these tests. Vulnerability scanning must run in CI: this workspace policy
denies the Go vulnerability database host.
