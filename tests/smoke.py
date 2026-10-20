#!/usr/bin/env python3
"""Real CLI + HTTP + PostgreSQL + Mailpit journey; no mocked services."""
import json
import os
import pathlib
import re
import secrets
import subprocess
import time
import urllib.request

root = pathlib.Path(__file__).resolve().parents[1]
run_id = secrets.token_hex(5)
base = os.environ.get("SINET_API", "http://127.0.0.1:8080")
mailpit = os.environ.get("MAILPIT_URL", "http://127.0.0.1:8025")


def get_json(url):
    with urllib.request.urlopen(url, timeout=5) as response:
        return json.load(response)


def cli(who, *args, stdin=None, ok=True):
    env = dict(os.environ, SINET_API=base,
               SINET_CONFIG=str(root / "work" / f"smoke-{run_id}-{who}.json"))
    result = subprocess.run([str(root / "bin" / "SINet"), *args],
                            env=env, input=stdin, text=True, capture_output=True, timeout=35)
    if (result.returncode == 0) != ok:
        raise AssertionError(f"command {args[0]} unexpected exit: {result.stderr}")
    print(f"PASS {args[0]} {' '.join(args[1:3])} ({'allowed' if ok else 'denied'})")
    if not ok:
        return result.stderr
    try:
        return json.loads(result.stdout)
    except json.JSONDecodeError:
        return result.stdout


def signin(who):
    email = f"{who}-{run_id}@example.test"
    cli(who, "login", "request", email)
    code = None
    for _ in range(30):
        messages = get_json(mailpit + "/api/v1/messages")["messages"]
        for message in messages:
            if any(address["Address"] == email for address in message["To"]):
                detail = get_json(mailpit + "/api/v1/message/" + message["ID"])
                match = re.search(r"Your code is (\d{8})", detail["Text"])
                if match:
                    code = match.group(1)
                    break
        if code:
            break
        time.sleep(0.1)
    assert code, "OTP not delivered to actual Mailpit SMTP sink"
    cli(who, "login", "verify", email, stdin=code + "\n")
    cli(who, "login", "verify", email, stdin=code + "\n", ok=False)
    cli(who, "claim", who + run_id)
    return who + run_id


def main():
    assert cli("alice", "health")["status"] == "ok"
    alice = signin("alice")
    bob = signin("bob")
    assert cli("alice", "me")["model_available"] is False
    cli("alice", "settings", "set", "bio", "A real CLI prototype journey")
    cli("alice", "settings", "set", "theme", "dark")
    assert cli("bob", "user", alice).get("email") is None
    assert cli("alice", "find", bob)["items"][0]["handle"] == bob
    cli("alice", "room", "lobby", "read", ok=False)
    for who in ("alice", "bob"):
        cli(who, "rooms")
        cli(who, "room", "lobby", "join")
    message = cli("alice", "room", "lobby", "say", "Hello from the real CLI")
    assert cli("bob", "room", "lobby", "read", "--limit", "1")["items"][0]["id"] == message["id"]
    cli("bob", "report", "message", str(message["id"]), "Smoke test moderation fixture")
    cli("bob", "block", alice)
    assert cli("bob", "room", "lobby", "read")["items"] == []
    cli("bob", "unblock", alice)
    note = cli("alice", "note", "new", "My first note", "Persisted in PostgreSQL")
    cli("bob", "note", "get", note["id"], ok=False)
    cli("alice", "note", "edit", note["id"], "Updated", "--stdin", stdin="Saved across CLI processes\n")
    assert cli("alice", "note", "get", note["id"])["body"] == "Saved across CLI processes\n"
    cli("alice", "plug", "validate", "connectors/examples/echo.json")
    cli("alice", "plug", "list")
    cli("alice", "plug", "info", "echo", "1.0.0")
    tool = ("agent", "call", "echo", "1.0.0", "repeat", '{"text":"Hello SINet"}')
    cli("alice", *tool, ok=False)
    cli("alice", "plug", "install", "echo", "1.0.0")
    cli("alice", *tool, ok=False)
    cli("alice", "plug", "grant", "echo", "1.0.0", "repeat", "--approve")
    result = cli("alice", *tool)
    assert result["mode"] == "deterministic_tool" and result["output"]["text"] == "Hello SINet"
    cli("bob", *tool, ok=False)
    assert cli("alice", "agent", "audit")["items"]
    cli("alice", "plug", "installed")
    cli("alice", "plug", "revoke", "echo", "repeat")
    cli("alice", *tool, ok=False)
    cli("alice", "plug", "remove", "echo")
    assert cli("alice", "agent", "ask", "Hello")["mode"] == "no_model"
    cli("alice", "agent", "usage")
    card = cli("alice", "share")
    assert alice in card and "/u/" + alice in card
    with urllib.request.urlopen(base + "/u/" + alice) as response:
        html = response.read().decode()
    assert "og:title" in html and "example.test" not in html
    exported = cli("alice", "export")
    assert exported["notes"]["items"][0]["id"] == note["id"]
    for who in ("alice", "bob"):
        cli(who, "delete-account", "--confirm-delete")
        cli(who, "me", ok=False)
    print("PASS full CLI/SMTP/PostgreSQL hero journey; two accounts; no credentials printed")


if __name__ == "__main__":
    main()
