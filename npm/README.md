# SINet prototype CLI

Your terminal. Your network.

```sh
npm install -g @ao3575911/sinet@0.1.1
SINet version
# Or run without a global installation:
npx --yes @ao3575911/sinet@0.1.1 version
```

Requires Node.js 18 or newer. Supports Linux x64, macOS arm64 and Windows x64.
This package bundles the Go CLI from the existing v0.1.1 GitHub release; it has
no dependencies, installation scripts or runtime binary downloads. It is a CLI,
not a JavaScript library. Native macOS and Windows execution is not yet validated.

SINet is a prototype and requires a running SINet API. The default API is
http://127.0.0.1:8080. See the [release and setup guide](https://github.com/ao3575911/sinet/blob/main/docs/RELEASE.md).
The API server, database and SMTP service are not included in this npm package.

Release archives are checked against pinned SHA-256 checksums during packaging.
For independent GitHub build provenance verification, follow the release guide.
