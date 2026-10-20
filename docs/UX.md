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
SINet wallet connect
```

The bare-path shorthand is optional and can never automatically upload files. The `--to` destination and confirmation are required for copy/move operations. Device availability and OS privileges can differ by platform.

## Onboarding

Welcome → `Email` → receive/enter OTP → handle availability check → transactional claim → dashboard with profile, agent, connector discovery, rooms → plug/install first example → perform permissioned tool call → share identity card.

## Dashboard layout

Profile/Handle | Agent | Plug-ins | People/Rooms | Notes | Settings

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
