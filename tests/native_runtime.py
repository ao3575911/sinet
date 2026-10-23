"""Run the native CLI against isolated real PostgreSQL and SMTP processes.

Requires initdb/pg_ctl/createdb on PATH, bin/mailpit and the Go toolchain.
Writes only nonsecret build evidence to work/native-artifacts after every check.
"""
import hashlib
import json
import os
from pathlib import Path
import platform
import subprocess
import sys
import tempfile
import time
import urllib.request

root = Path(__file__).resolve().parents[1]
os.chdir(root)
suffix = ".exe" if os.name == "nt" else ""
env = dict(os.environ, TEST_DATABASE_URL="postgres://sinet@127.0.0.1:15432/sinet?sslmode=disable",
           DATABASE_URL="postgres://sinet@127.0.0.1:15432/sinet?sslmode=disable",
           APP_HTTP_ADDR="127.0.0.1:18080", APP_PUBLIC_URL="http://127.0.0.1:18080",
           SINET_API="http://127.0.0.1:18080", SMTP_ADDR="127.0.0.1:11025",
           MAILPIT_URL="http://127.0.0.1:18025")


def run(*args, **kwargs):
    subprocess.run(args, env=env, check=True, **kwargs)


def ready(url):
    for _ in range(100):
        try:
            with urllib.request.urlopen(url, timeout=1):
                return
        except OSError:
            time.sleep(0.2)
    raise AssertionError("native fixture failed to become ready")


def main():
    expected = "amd64" if os.name == "nt" else "arm64"
    actual = subprocess.check_output(["go", "env", "GOARCH"], text=True).strip()
    assert platform.system() in ("Darwin", "Windows") and actual == expected, "unexpected native runner architecture"
    (root / "work").mkdir(exist_ok=True)
    (root / "bin").mkdir(exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="sinet-native-") as temp:
        if os.name == "nt":
            # initdb drops the elevated Administrators SID. Python's private
            # temp directory may be owned by that group on a hosted runner;
            # give the actual user explicit access before the token is reduced.
            user = subprocess.check_output(["whoami"], text=True).strip()
            run("icacls", temp, "/grant:r", user + ":(OI)(CI)F")
        pg = str(Path(temp) / "pg")
        run("initdb", "-D", pg, "-U", "sinet", "-A", "trust", "--encoding=UTF8", "--no-locale")
        run("pg_ctl", "-D", pg, "-l", str(Path(temp) / "postgres.log"),
            "-o", "-h 127.0.0.1 -p 15432", "-w", "start")
        children = []
        try:
            run("createdb", "-h", "127.0.0.1", "-p", "15432", "-U", "sinet", "sinet")
            run("go", "test", "-race", "./...", "-count=1")
            # Match release compiler flags; test the actual CGO-disabled candidate.
            for component in ("SINet", "server"):
                subprocess.run(["go", "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w",
                                "-o", f"bin/{component}{suffix}", f"./cmd/{component}"],
                               env=dict(env, CGO_ENABLED="0"), check=True)
            for executable, args in (("mailpit", ["--listen", "127.0.0.1:18025", "--smtp", "127.0.0.1:11025"]),
                                     ("server", [])):
                log = open(Path(temp) / (executable + ".log"), "wb")
                try:
                    children.append(subprocess.Popen([str(root / "bin" / (executable + suffix)), *args],
                                                     env=env, stdout=log, stderr=log))
                finally:
                    log.close()
            ready(env["SINET_API"] + "/healthz")
            ready(env["MAILPIT_URL"] + "/api/v1/messages")
            run(sys.executable, "tests/smoke.py")
            run(sys.executable, "tests/tui_smoke.py")
            # No fixture accounts or isolated test schemas may remain.
            count = subprocess.check_output(["psql", "-h", "127.0.0.1", "-p", "15432", "-U", "sinet", "-d", "sinet", "-Atc",
                "SELECT (SELECT count(*) FROM accounts) + (SELECT count(*) FROM information_schema.schemata WHERE schema_name LIKE 'test_%')"], text=True)
            assert count.strip() == "0", "native fixtures were not cleaned up"
            output = root / "work" / "native-artifacts"
            output.mkdir(exist_ok=True)
            target = "windows-amd64" if os.name == "nt" else "darwin-arm64"
            name = "SINet-" + target + suffix
            binary = (root / "bin" / ("SINet" + suffix)).read_bytes()
            (output / name).write_bytes(binary)
            digest = hashlib.sha256(binary).hexdigest()
            (output / "SHA256SUMS").write_text(f"{digest}  {name}\n", encoding="utf-8")
            evidence = {"source_commit": subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip(),
                        "platform": platform.platform(), "machine": platform.machine(),
                        "compiler": subprocess.check_output(["go", "version"], text=True).strip(),
                        "sha256": digest, "checks": ["race suite", "OS credential store", "CLI SMTP PostgreSQL", "native PTY", "fixture cleanup"],
                        "status": "tested candidate; not a signed production release"}
            (output / "VALIDATION.json").write_text(json.dumps(evidence, indent=2) + "\n", encoding="utf-8")
            print("PASS native runtime; verified SHA256", digest)
        finally:
            for child in reversed(children):
                child.terminate()
                child.wait(timeout=15)
            run("pg_ctl", "-D", pg, "-m", "immediate", "-w", "stop")


if __name__ == "__main__":
    main()
