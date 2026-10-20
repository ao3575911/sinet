# SINet project context

Source-linked synthesis of materials supplied on 20 October 2026. This summary is a
navigation aid; consult the original documents for exact requirements. The user's current
request determines the work to perform. No application has been implemented by this import.

## Product and positioning

SINet means **Super Intelligence Network** in the repository README, whose existing line is
“Intelligence is networked.” The supplied concept describes a terminal-native social
computing platform combining identity, communities, a private personal agent, and an
extensible connector marketplace. Target users include social explorers, terminal newcomers,
builders/connector creators, power users, and later business operators.

- Product positioning: “Your username. Your agent. Your terminal. Your network.”
- Brand tagline: “Your terminal. Your network.”
- Product thesis: make the username desirable, the terminal approachable, and connectors indispensable.
- Hero loop: claim a handle, meet people, install a connector, perform a useful action
  through the personal agent, and share the result.

Sources: [concept](originals/SINet-Concept.txt), [starter PRD](extracted/cli-social-network-starter/docs/PRD.md),
[personas](extracted/cli-social-network-starter/docs/PERSONAS.md), [brand README](extracted/SINet-Brand-Kit/README.md).

## Proposed MVP and technical direction

The detailed starter narrows the broad concept to a verifiable vertical slice: CLI/TUI
onboarding, email OTP, an atomic unique normalized handle, public profile/share card and
minimal web profile, public directory/community room with reporting and blocking, private
notes/settings, a personal agent, and a reviewed connector registry with install/uninstall,
per-operation grants, revocation, and audit records. An `echo` connector provides a safe
deterministic first example. A real configured model is distinct from deterministic tools.

- CLI and API: Go; Bubble Tea where useful; scriptable commands; Linux/macOS/Windows.
- Persistence: PostgreSQL, transactional constraints, versioned migrations.
- Development email: Mailpit; production mail needs an explicitly configured adapter.
- AI: provider-independent adapter with an OpenAI-compatible endpoint, quotas, and an honest
  unavailable state when no provider is configured.
- Connectors: JSON manifests/schema, reviewed built-ins and vetted HTTPS services;
  publisher submissions begin in review; versions and grants are enforced server-side.
- Proposed packaging: Docker Compose for API/Postgres/Mailpit and portable CLI binaries.
- Public API draft: REST/JSON under `/v1`, protected bearer/session authentication.

Sources: [build plan](originals/SINet-Build-Plan-PRD.txt), [architecture](extracted/cli-social-network-starter/docs/ARCHITECTURE.md),
[API](extracted/cli-social-network-starter/docs/API.md), [ADR 0001](extracted/cli-social-network-starter/docs/ADR/0001-runtime-and-scope.md),
[connector schema](extracted/cli-social-network-starter/contracts/connector.schema.json).
ADR 0001 is marked proposed; these describe intended architecture, not installed components.

## Security and truthful capability boundaries

The supplied security model calls for expiring one-use hashed OTPs, rate limits, atomic
handle claims, revocable sessions, tenant isolation, private email/notes, and no secret
logging. Connector execution defaults to denied and requires account/connector/operation
grants, consent for mutation or external disclosure, validated arguments, bounded output,
and audit records. Remote HTTPS connectors require verified ownership and SSRF defenses
including DNS and connected-IP checks, blocked private networks, and restricted redirects.
Model or third-party text cannot grant permissions.

The MVP excludes uploaded arbitrary code execution, unrestricted hosted shells, cloud-agent
access to the user's host or USB, crypto custody, and claims of a working full email inbox.
An email alias requires a controlled domain and routing; ENS contact metadata does not
provide an SMTP inbox. Wallets/ENS/DIDs, business identities/commerce, local bridges, isolated
cloud compute, remote MCP adapters, and mobile are later tracks.

Sources: [security](extracted/cli-social-network-starter/docs/SECURITY.md), [PRD](extracted/cli-social-network-starter/docs/PRD.md),
[runtime agent policy reference](extracted/cli-social-network-starter/prompts/CLOUD_AGENT_SYSTEM.md).

## Milestones and evidence

The starter roadmap proposes M0 runnable skeleton; M1 identity/activation; M2 social/notes;
M3 connector hero loop; M4 real configured model plus hardening. Its acceptance matrix
requires a clean bootstrap and actual tests of OTP replay/expiry, concurrent handle claims,
tenant isolation, room authorization, manifest review, connector grants/revocation, private
network egress denial, quotas, and honest no-model behavior. These are planned checks and
have not been run against an application because the supplied starter contains specs only.

Sources: [roadmap](extracted/cli-social-network-starter/docs/ROADMAP.md),
[test matrix](extracted/cli-social-network-starter/docs/TESTS.md).

## Brand and assets

The brand is sharp, optimistic, modular, credible, and social. Voice is concise, welcoming,
and technically competent, without elitist jargon, false urgency, or empty AI promises.
The visual system uses a ribbon-like S mark, cyan/blue/violet gradients, dark terminal
surfaces, and light or monochrome alternatives. SVG marks have outlined lettering.

| Token | Hex |
| --- | --- |
| Midnight | `#080F19` |
| Deep | `#060A12` |
| Surface | `#101928` |
| Signal cyan | `#00D1FF` |
| Signal blue | `#1E5AB8` |
| Agent violet | `#885CF6` |
| Ice | `#F8FAFC` |
| Secondary | `#B8C6D8` |
| Muted/slate | `#64748B` |

Use Inter Display/Inter for UI/headings, system monospace for commands/metadata. Font files
are not included. Use the supplied dark/light variants on appropriate backgrounds, preserve
geometry and at least 15% symbol-height clearspace, and maintain accessible contrast.
Assets include SVG/PNG logos, favicons/app icons, social exports, wallpapers/patterns,
CSS/JSON tokens, terminal splash, email signature, launch copy, moodboard, and PDF guidelines.
The kit is a concept identity; trademark/domain clearance is not established.

Sources: [brand README](extracted/SINet-Brand-Kit/README.md), [tokens](extracted/SINet-Brand-Kit/tokens/colors.json),
[integration guide](extracted/SINet-Brand-Kit/extras/README-INTEGRATION.md),
[PDF text](derived/SINet-Brand-Guidelines.txt), [visual overview](extracted/SINet-Brand-Kit/previews/brand-overview.jpg).

## Differences to preserve for later decisions

- Several documents still use `project-name` and placeholder domains; the brand kit uses
  `SINet` and launch copy uses `sinet`. Original spelling is preserved rather than rewritten.
- Broad concept features extend beyond the starter's narrowed MVP. Do not silently promote
  later-stage features into launch commitments.
- The detailed PRD includes a minimal shareable web profile in P0 and a broader public web
  viewer/mobile-friendly companion in P1; determine intended extent during implementation.
- Product positioning and the short brand tagline coexist; they serve different copy contexts.
- Historical build prompts describe implementation work and commits. Their presence does
  not turn the user's storage/context request into authorization to start that work.

Preserve these sources and document future resolutions separately.
