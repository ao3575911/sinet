#!/usr/bin/env python3
"""POSIX PTY acceptance: guided signup, arrow navigation and persistent note editor."""
import errno
import json
import os
import pathlib
import pty
import re
import secrets
import select
import subprocess
import time
import urllib.request

root = pathlib.Path(__file__).resolve().parents[1]
run_id = secrets.token_hex(5)
email = f"tui-{run_id}@example.test"
env = dict(os.environ, SINET_CONFIG=str(root / "work" / f"tui-{run_id}.json"))
master, slave = pty.openpty()
proc = subprocess.Popen([str(root / "bin" / "SINet")], stdin=slave, stdout=slave, stderr=slave, env=env)
os.close(slave)
seen = b""


def expect(text):
    global seen
    target = text.encode()
    deadline = time.monotonic() + 15
    while target not in seen:
        if time.monotonic() > deadline:
            raise AssertionError(f"TUI timeout waiting for {text!r}")
        ready, _, _ = select.select([master], [], [], 0.2)
        if ready:
            chunk = os.read(master, 8192)
            if not chunk:
                raise AssertionError("TUI ended early")
            seen += chunk
    before, seen = seen.split(target, 1)


def send(text):
    os.write(master, text.encode())


def get_json(path):
    with urllib.request.urlopen(os.environ.get("MAILPIT_URL", "http://127.0.0.1:8025") + path, timeout=5) as response:
        return json.load(response)


try:
    expect("Sign in with email")
    send("1")
    expect("Email:")
    send(email + "\n")
    expect("Sign-in code (hidden):")
    code = None
    for _ in range(30):
        for message in get_json("/api/v1/messages")["messages"]:
            if any(address["Address"] == email for address in message["To"]):
                detail = get_json("/api/v1/message/" + message["ID"])
                code = re.search(r"Your code is (\d{8})", detail["Text"]).group(1)
                break
        if code:
            break
        time.sleep(0.1)
    assert code
    send(code + "\n")
    expect("Choose your handle")
    send("tui" + run_id + "\n")
    expect("Your network")
    send("4")
    expect("Note ID to edit")
    send("\n")
    expect("Title:")
    send("TUI note\n")
    expect("A single . saves")
    send("Created in the built-in editor\n.\n")
    expect("Your network")
    send("8")
    expect("\r\nWallet, email and services\r\n")
    send("2")
    expect("Compose a draft")
    send("2")
    expect("Recipient (empty is allowed in a draft):")
    send("friend@example.test\n")
    expect("Subject:")
    send("TUI email draft\n")
    expect("A single . saves")
    send("Private email draft from the terminal\n.\n")
    expect("Your network")
    drafts_result = subprocess.run([str(root / "bin" / "SINet"), "email", "list"], env=env, text=True, capture_output=True, check=True)
    drafts = json.loads(drafts_result.stdout)["items"]
    assert drafts[0]["subject"] == "TUI email draft" and drafts[0]["status"] == "draft"
    send("8")
    expect("\r\nWallet, email and services\r\n")
    send("2")
    expect("Review a draft in Gmail")
    send("3")
    expect("Draft ID:")
    send(drafts[0]["id"] + "\n")
    expect("Type gmail to prepare it:")
    send("gmail\n")
    expect("mail.google.com")
    expect("Your network")
    send("5")
    expect("GitHub public facts and audit")
    send("8")
    expect("GitHub public: ao3575911/sinet")
    send("4")
    expect("operator must deploy and approve it first")
    expect("Your network")
    send("\x1b[B")
    expect("> 2. Find people")
    send("q")
    proc.wait(timeout=5)
    assert proc.returncode == 0
    result = subprocess.run([str(root / "bin" / "SINet"), "note", "list"], env=env, text=True, capture_output=True, check=True)
    notes = json.loads(result.stdout)["items"]
    assert notes[0]["title"] == "TUI note" and "built-in editor" in notes[0]["body"]
    subprocess.run([str(root / "bin" / "SINet"), "delete-account", "--confirm-delete"], env=env, capture_output=True, check=True)
    print("PASS PTY: signup, handle, note editor, private email draft, explicit Gmail handoff, unapproved GitHub gate, arrow navigation, q/back")
finally:
    if proc.poll() is None:
        proc.kill()
        proc.wait()
    os.close(master)
    # Clean up this run even if a navigation assertion fails.
    if pathlib.Path(env["SINET_CONFIG"]).exists():
        cleanup = subprocess.run([str(root / "bin" / "SINet"), "delete-account", "--confirm-delete"], env=env, capture_output=True)
        if cleanup.returncode:
            raise AssertionError("PTY fixture account cleanup failed")
