# Contributing and community

Use labelled issues and PRs with clear acceptance and command/test evidence. Keep Go
dependencies lean. Sign commits, verify them and register the signing key with GitHub.
Preserve reference materials and migration checksums. Record architecture assumptions.

```sh
make check
make integration
make smoke
make smoke-tui
make cross
```

Database tests must not be skipped when claiming acceptance. Smoke tests require real API,
PostgreSQL and Mailpit. Changes to grants, egress or data disclosure need explicit security
review. Read [CONTRIBUTING.md](https://github.com/ao3575911/sinet/blob/main/CONTRIBUTING.md)
and [community expectations](https://github.com/ao3575911/sinet/blob/main/CODE_OF_CONDUCT.md).

[Documentation home](Home.md)
