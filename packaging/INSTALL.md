# SINet CLI

Your terminal. Your network. This package runs the SINet Go CLI using a bundled executable.
It is a production-oriented prototype, not a public hosted service or a production launch.

```sh
npx @ao3575911/sinet version
# Or install the command:
npm install --global @ao3575911/sinet
SINet version
# Python alternative (use a virtual environment where required):
python3 -m pip install sinet-cli
SINet version
```

Run `SINet` for guided sign-in, or `SINet help` for scriptable commands. Network commands
require a running API; version/help run locally. Start the development stack from the
[source repository](https://github.com/ao3575911/sinet), with development email in Mailpit.
`SINET_API` selects the API origin. The agent never executes uploaded plugin code on your host.

Supported binaries: Linux x64, macOS arm64, Windows x64. npm requires Node.js 22+; pip requires
Python 3.9+ and a compatible platform wheel. macOS wheels require macOS 14+; Linux wheels
target manylinux_2_28_x86_64. No Linux ARM or macOS Intel executable is included. macOS native
validation remains outstanding; Windows/Linux installer checks are recorded in repository CI.

There are no install hooks, runtime downloads or additional runtime libraries. Both launchers
verify the bundled executable's SHA-256 before running it. BUILDINFO.json identifies the
original Go release source/compiler. GitHub build attestations and registry provenance
provide distribution evidence; local hashes alone do not authenticate an altered package.
These binaries are not Apple-notarized or Windows Authenticode-signed.

License: Apache-2.0. See LICENSE and THIRD_PARTY_NOTICES.md. Full installation, operating,
security and known-limit documentation: https://github.com/ao3575911/sinet.
