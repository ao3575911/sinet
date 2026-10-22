#!/usr/bin/env python3
"""Local Chromium test with an explicit EIP-1193 fixture, not a live wallet demo.

Requires a running development API/Mailpit, built CLI, playwright and Chromium.
No Google, chain, wallet-provider or other external requests are permitted.
"""
import json
import os
import pathlib
import re
import secrets
import shutil
import subprocess
import time
import urllib.parse
import urllib.request

from playwright.sync_api import sync_playwright

root = pathlib.Path(__file__).resolve().parents[1]
run_id = secrets.token_hex(5)
email = f"wallet-browser-{run_id}@example.test"
env = dict(os.environ, SINET_CONFIG=str(root / "work" / f"wallet-browser-{run_id}.json"))
api = os.environ.get("SINET_API", "http://127.0.0.1:8080")
mailpit = os.environ.get("MAILPIT_URL", "http://127.0.0.1:8025")
vector = json.loads((root / "internal/wallet/testdata/personal_sign.json").read_text())


def cli(*args, input=None):
    result = subprocess.run([str(root / "bin/SINet"), *args], env=env, input=input, text=True, capture_output=True, check=True)
    return json.loads(result.stdout)


def get_json(url):
    with urllib.request.urlopen(url, timeout=5) as response:
        return json.load(response)


signed_in = False
try:
    cli("login", "request", email)
    code = None
    for _ in range(30):
        for message in get_json(mailpit + "/api/v1/messages")["messages"]:
            if any(address["Address"] == email for address in message["To"]):
                detail = get_json(mailpit + "/api/v1/message/" + message["ID"])
                code = re.search(r"Your code is (\d{8})", detail["Text"]).group(1)
                break
        if code:
            break
        time.sleep(0.1)
    assert code
    cli("login", "verify", email, input=code + "\n")
    signed_in = True
    handle = "wb" + run_id
    cli("claim", handle)
    challenge = cli("wallet", "challenge", vector["address"])
    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(executable_path=os.environ.get("CHROMIUM_PATH", shutil.which("chromium")), headless=True, args=["--no-sandbox"])
        context = browser.new_context()
        requests = []

        def route_request(route):
            requests.append(route.request.url)
            if route.request.url == api + "/wallet/link":
                route.continue_()
            else:
                route.abort()

        context.route("**/*", route_request)
        context.add_init_script("""window.walletCalls=[];window.walletMode='valid';
          window.ethereum={request:async(request)=>{
            window.walletCalls.push(request);
            if(window.walletMode==='null_error')throw null;
            if(request.method==='eth_requestAccounts')return [window.walletMode==='wrong_address'?'0x1111111111111111111111111111111111111111':%s];
            if(request.method==='eth_chainId')return window.walletMode==='wrong_chain'?'0xa':'0x1';
            if(request.method==='personal_sign')return %s;
            throw new Error('unexpected wallet operation');
          }};""" % (json.dumps(vector["address"]), json.dumps(vector["signature"])))
        for mode in ("valid", "wrong_address", "wrong_chain", "null_error"):
            page = context.new_page()
            page.goto(challenge["signing_url"])
            assert page.locator("#message").inner_text() == challenge["message"]
            assert "@" + handle in page.locator("#account").inner_text()
            assert page.evaluate("location.hash") == ""
            page.evaluate("mode => window.walletMode=mode", mode)
            page.locator("#sign").click()
            page.wait_for_function("document.getElementById('status').textContent.length>0")
            calls = page.evaluate("window.walletCalls")
            if mode == "valid":
                assert page.locator("#signature").input_value() == vector["signature"]
                assert [call["method"] for call in calls] == ["eth_requestAccounts", "eth_chainId", "personal_sign"]
                assert calls[-1]["params"] == ["0x" + challenge["message"].encode().hex(), vector["address"]]
            else:
                assert page.locator("#signature").input_value() == ""
                assert all(call["method"] != "personal_sign" for call in calls)
            if mode == "null_error":
                assert page.locator("#status").inner_text() == "Wallet signing was cancelled or unavailable. You can try again."
                assert page.locator("#sign").is_enabled()
            page.close()
        params = urllib.parse.parse_qs(urllib.parse.urlsplit(challenge["signing_url"]).fragment)
        params["message"] = [params["message"][0].replace("No transactions are authorized.", "Send funds now.")]
        hostile = api + "/wallet/link#" + urllib.parse.urlencode(params, doseq=True)
        page = context.new_page()
        page.goto(hostile)
        assert page.locator("#sign").is_disabled()
        assert page.evaluate("window.walletCalls") == []
        page.close()
        assert requests and all(url == api + "/wallet/link" for url in requests)
        browser.close()
    print("PASS browser wallet fixture: exact bytes/account/origin, fragment cleared, wrong address/chain/message denied, null rejection handled, no external requests")
finally:
    if signed_in:
        cli("delete-account", "--confirm-delete")
