# ADR 0005: wallet ownership linking and Gmail draft handoff

Accepted for this unreleased implementation slice, 20 October 2026. The owner
requested wallet and email first, then the other sections, and chose Gmail as
the initial email provider. This extends the later-track scope without adopting
custody, transfers, mailbox sync or a hosted inbox at signup.

## Wallet

Support Ethereum EOAs on chain ID 1. A signed-in, activated SINet account requests
a five-minute SIWE-format challenge naming its handle, site origin, address,
purpose, nonce, issue/expiration times and chain. The stored challenge is bound
to that account and current configured origin. Verify the exact EIP-191
personal_sign bytes with secp256k1 public-key recovery, EIP-55 address checks and
canonical low-s signatures. Allow v=0/1/27/28. No wallet-generated session or
wallet login replaces the existing email OTP authentication.

Only the newest challenge per account is active. Consume successful proofs once;
five invalid attempts exhaust a challenge. Limit challenge issuance to five per
15 minutes and links to five per account. Globally unique chain/address links
prevent silent reassignment. Require a fresh proof after unlinking. Persist only
the address, chain, verification time and message hash; signatures and wallet
secrets are not stored. Links/challenges are private and cascade on account deletion.

Use the audited Decred secp256k1 implementation v4.4.1 and x/crypto v0.58.0 for
legacy Keccak, rather than implementing curve math. Verify compatibility against
an independent ethers 6.15.0 public signature fixture. Runtime has no ethers
dependency, RPC endpoint, blockchain writes, balances or contract-wallet support.

A small nonce-CSP browser helper invokes only eth_requestAccounts, eth_chainId
and personal_sign. It shows the exact SINet handle/address/message. Challenge
data travels in a URL fragment and is removed from browser history on load;
it is not placed in query strings. The helper rejects wrong origins, accounts,
chains, expired or changed-purpose messages. It makes no application network
requests and displays the signature for manual CLI completion. The authenticated
CLI remains necessary to submit the proof. No wallet private key is requested.

## Email

Gmail means an existing user-controlled Gmail account; Google controls gmail.com.
Private drafts are real database records, scoped to their account and included
in export/deletion. Enforce body/header size limits, recipient validation, safe
text, pagination and at most 100 drafts. They stay drafts in SINet.

With explicit user approval, the CLI encodes a fixed mail.google.com compose
URL locally. Opening it shares the draft with Google; the user reviews the sender,
recipients and body and sends inside Gmail. No mail is sent by SINet and no
mailbox access is implied. Compose URLs are capped at 8192 bytes; larger drafts
can be edited/copied into Gmail. A fixed external Gmail inbox link is available.

Automatic synchronization and sending require a separately configured Google
OAuth application, per-account consent and encrypted token storage. Platform
aliases require an owned domain, routing and verified provisioning. Neither is
configured by this slice. Existing OTP SMTP remains an authentication service,
not a mailbox or arbitrary-send service. Assistant connector credentials are not
reused as application credentials.

## Navigation and release

Keep existing dashboard choices 1–7; choice 8 opens wallet/email/services. Other
sections expose accurate plans and setup steps for business, ENS/DID identity,
files/devices, terminal/cloud and mobile. These screens do not activate those
capabilities. Forward migration 0005 adds only wallet and draft tables; prior
migration bytes remain unchanged. Target release is 0.2.0, not an overwritten
0.1.1 npm/PyPI package. The owner authorized delivery through a signed squash commit to main.
