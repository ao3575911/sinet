# SINet - Brand identity kit

**Positioning:** A terminal-native social network connecting people, AI agents, and plugins.  
**Tagline:** **Your terminal. Your network.**  
**Brand character:** Sharp, optimistic, modular, credible and social.  
**Version:** 1.0 (concept identity, subject to name/trademark clearance).

## Package contents

- `logos/svg/` - transparent, scalable, outlined brand marks and lockups (recommended source).
- `logos/png/` - ready-to-use transparent logo exports in multiple sizes.
- `icons/` - dark, light and monochrome app icons; favicon, iOS icon, PWA icons and manifest.
- `social/` - profile avatar, X, LinkedIn, YouTube, OpenGraph, GitHub, feed square and story images.
- `backgrounds/` - desktop/mobile wallpapers, network pattern and terminal grid.
- `docs/SINet-Brand-Guidelines.pdf` - concise visual identity guide.
- `tokens/` - CSS custom properties and machine-readable JSON design tokens.
- `extras/` - launch copy, terminal splash, email signature example, implementation guidance.
- `previews/brand-overview.jpg` - overview built directly from the production assets.

## Brand colors

| Color | Hex | Use |
|---|---|---|
| Midnight | `#080F19` | Primary dark background |
| Deep | `#060A12` | Near-black interface backgrounds |
| Cyan | `#00D1FF` | Active states, cursor, calls to action |
| Deep blue | `#1E5AB8` | Supporting graphics |
| Violet | `#885CF6` | Agents, intelligence, identity |
| Ice | `#F8FAFC` | Primary text |
| Slate | `#64748B` | Secondary/disabled text (not for small text on midnight) |

## Typography

**Inter Display / Inter** for wordmarks, headings and UI. **System monospace** for commands, metadata and console elements. The wordmark is converted to scalable vector outlines, so no font dependency is needed in the logo files.

## Correct usage

Use `primary-dark` on dark surfaces and `primary-light` on light surfaces. Keep a minimum clearspace equal to 15% of symbol height. Use the monochrome mark when gradients do not reproduce correctly. Never treat the icon as an indication of access to privileged machine resources.

## Web integration

```html
<link rel="icon" href="/icons/favicon.ico" sizes="any">
<link rel="apple-touch-icon" href="/icons/apple-touch-icon.png">
<link rel="manifest" href="/icons/site.webmanifest">
<meta name="theme-color" content="#080F19">
```

## Rights / limitations

Original design prepared for this project, not a trademark or domain availability opinion. Clear **SINet** before commercial use. No third-party font files are included. The concept moodboard is separate from these production SVGs.
