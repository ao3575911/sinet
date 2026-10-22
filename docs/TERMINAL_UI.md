# Terminal dashboard and consistent launches

Source candidate **0.2.0-rc.2** includes Wallet, Email and Services at **8** in the
nine-choice dashboard. Published npm/PyPI **0.1.1** still contains the older menu.
Working-directory differences usually mean different executables, not different
account settings. `./SINet` uses a file in the current directory; `SINet` uses PATH.

Build and install a reviewed checkout with the supported Go toolchain:

```sh
make install
hash -r
SINet version
SINet doctor
```

The installer places one binary in `~/.local/share/SINet/bin/SINet` and makes
`~/.local/bin/SINet` point to it. Existing launchers are backed up beside the entry
point; rebuilding a repository binary alone does not update an installed package.
Add `~/.local/bin` to PATH if absent. Restart shells with cached command paths.
The installer does not download code or change a registry package version.

An explicitly identified extra launcher can use the same binary:

```sh
python3 scripts/install_local.py --entry-point /absolute/path/to/another/SINet
```

Only use paths you intend to replace. Root-owned files need the owner's intervention;
no privileged writes happen automatically. `doctor` reports the executing binary,
compiled version, checksum and whether PATH launchers contain the same binary.
It inspects files without executing PATH candidates and prints no session credentials.
Pinned `npx`/pipx package invocations stay at their pinned release until explicitly
upgraded after publication. Do not overwrite or move historical release tags.

## Appearance and navigation

Run `SINet ui` offline, or select **7 Settings → Appearance** after signing in.
Choose dark (cyan) or light (blue), automatic/always/no colour, and comfortable or
compact layout. Preferences apply immediately and persist locally in
`~/.config/SINet/ui.json` on Linux (the OS configuration directory on other systems).
`SINET_UI_CONFIG` overrides that path. `SINet ui show` prints JSON; `SINet ui reset`
recovers from invalid settings. Session and appearance files are separate.
`NO_COLOR` overrides the colour preference. For screen readers or basic terminals,
`TERM=dumb` uses numbered line prompts; redirected output contains no screen controls.

Arrow keys, j/k or Tab move the highlight; Enter selects. Digits select directly.
Home/End jump to the first/last option, ? toggles help and q/Esc goes back.
Ctrl-C exits a menu safely. The dashboard stays in place while navigating;
small terminals scroll the selection inside the menu. Menus render inline so
results and exact approval disclosures remain in terminal scrollback. Bracketed
paste cannot select an action. Plain command output remains scriptable JSON.

Settings also offers profile name/bio, public/private visibility, notifications,
account settings and launcher diagnostics. Wallet proofs, Gmail handoff and external
connector invocations retain their explicit authorization prompts. UI preferences
are not permission grants. Availability continues to depend on server setup.

## Verification

`tests/dashboard_pty.py` runs the built binary through a real PTY and a VT screen
emulator. It uses an isolated synthetic API, not live service evidence. Install the
pinned test dependency `pyte==0.8.2` and run it after `make build`.
`tests/tui_smoke.py` covers actual development API/SMTP signup, notes, email drafts
and the unapproved GitHub gate. See the validation record for observed results.
