# CLI command and journey contract

## Primary commands

```text
SINet                   # TUI welcome / sign-in / dashboard
SINet login            # email OTP sign-in
SINet claim <handle>   # reserve after verification
SINet me               # own profile and resources
SINet user <handle>    # public profile
SINet find <query>     # directory search
SINet rooms            # list public rooms
SINet room <slug>      # join/open a room
SINet note             # launch simple built-in text editor
SINet agent            # interact with agent
SINet plug list        # curated and community manifests
SINet plug info <id>   # inspect capabilities, publisher, privacy
SINet plug install <id>
SINet plug grant <id> <operation>
SINet plug remove <id>
SINet settings
SINet wallet connect ADDRESS # unreleased: ownership linking
SINet email                  # unreleased: private drafts and Gmail handoff
SINet help
```

## Illustrative, not shipped in P0

```text
SINet terminal --local
SINet terminal --cloud
SINet usb list
SINet usb copy ./report.pdf --to usb:WORK
SINet ./report.pdf --to usb:WORK
SINet business create <handle>
```

The bare-path shorthand is optional and can never automatically upload files. The `--to` destination and confirmation are required for copy/move operations. Device availability and OS privileges can differ by platform.

## Onboarding

Welcome → `Email` → receive/enter OTP → handle availability check → transactional claim → dashboard with profile, agent, connector discovery, rooms → plug/install first example → perform permissioned tool call → share identity card.

## Dashboard layout

Profile/Handle | Agent | Plug-ins | People/Rooms | Notes | Settings

Unreleased wallet/email expansion: dashboard item 8 opens Wallet, Email, Business,
ENS/external identity, Files/devices, Terminal/cloud, Mobile and Help. Wallet ownership
linking and private email drafts/Gmail handoff are functional; the other new sections
show accurate planned/setup states. See [commands and boundaries](WALLET_EMAIL.md).

Visible badges may indicate permissions, provenance and trial limits, not manufactured user scarcity. TUI always provides `?` help, escape/back, accessible terminal colors, and plain-text fallback.

## Implemented handle and interface rules

The executable and display brand are `SINet`, as specified by the owner. Handles
normalize by trimming surrounding whitespace and lowercasing, then require
`^[a-z][a-z0-9_]{2,23}$`. Unicode handles and reassignment are unavailable. Reserved names
are compared with underscores removed: admin, administrator, root, system, support, help,
security, official, sinet, api, www, mail, email, postmaster,
abuse, moderator, staff, null, undefined, lobby, agent, connectors, settings, login,
logout, me. Deleted handles remain reserved. Availability is advisory; the database
transaction decides the claim.

The dashboard is a plain keyboard menu with numbers, arrows/j/k, Enter, help and q/Escape.
The note editor replaces note text, saves on a single `.`, and cancels on `.cancel`.
Theme and notification settings are persisted preferences; this prototype uses plain
terminal rendering and does not send notifications. List commands are paginated rather
than live streamed. `room lobby` joins and reads; posting remains explicit.

## Public GitHub facts (0.2.0-rc.1 source)

Dashboard: Connectors → GitHub public facts and audit. Inspect/install/grant are
separate from the exact one-use invocation disclosure; type invoke to approve.
CLI: `SINet agent github` previews both recipients and fixed arguments, then
`SINet agent github --approve` obtains one consent and calls through the personal
agent. View `SINet agent audit`; revoke with `SINet plug revoke github-public
repository_summary`. Live service awaits owned hosting and operator review.
[Complete journey](GITHUB_CONNECTOR.md).
