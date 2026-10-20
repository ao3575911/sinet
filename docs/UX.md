# CLI command and journey contract

## Primary commands

```text
project-name                   # TUI welcome / sign-in / dashboard
project-name login            # email OTP sign-in
project-name claim <handle>   # reserve after verification
project-name me               # own profile and resources
project-name user <handle>    # public profile
project-name find <query>     # directory search
project-name rooms            # list public rooms
project-name room <slug>      # join/open a room
project-name note             # launch simple built-in text editor
project-name agent            # interact with agent
project-name plug list        # curated and community manifests
project-name plug info <id>   # inspect capabilities, publisher, privacy
project-name plug install <id>
project-name plug grant <id> <operation>
project-name plug remove <id>
project-name settings
project-name help
```

## Illustrative, not shipped in P0

```text
project-name terminal --local
project-name terminal --cloud
project-name usb list
project-name usb copy ./report.pdf --to usb:WORK
project-name ./report.pdf --to usb:WORK
project-name business create <handle>
project-name wallet connect
```

The bare-path shorthand is optional and can never automatically upload files. The `--to` destination and confirmation are required for copy/move operations. Device availability and OS privileges can differ by platform.

## Onboarding

Welcome → `Email` → receive/enter OTP → handle availability check → transactional claim → dashboard with profile, agent, connector discovery, rooms → plug/install first example → perform permissioned tool call → share identity card.

## Dashboard layout

Profile/Handle | Agent | Plug-ins | People/Rooms | Notes | Settings

Visible badges may indicate permissions, provenance and trial limits, not manufactured user scarcity. TUI always provides `?` help, escape/back, accessible terminal colors, and plain-text fallback.
