#!/usr/bin/env python3
"""Publish versioned wiki pages with the repository's configured Git signing identity."""
import pathlib
import re
import subprocess

root = pathlib.Path(__file__).resolve().parents[1]
wiki = root / 'work' / 'wiki'
remote = 'https://github.com/ao3575911/sinet.wiki.git'


def git(*args, cwd=wiki):
    return subprocess.check_output(['git', *args], cwd=cwd, text=True).strip()


wiki.parent.mkdir(exist_ok=True)
if not wiki.exists():
    subprocess.run(['git', 'clone', remote, str(wiki)], check=True)
if git('remote', 'get-url', 'origin') != remote:
    raise SystemExit('Unexpected wiki remote; refusing to publish')
for key in ['user.name', 'user.email', 'gpg.format', 'user.signingkey', 'commit.gpgsign', 'gpg.ssh.allowedSignersFile']:
    value = git('config', '--get', key, cwd=root)
    if not value:
        raise SystemExit('Missing signing configuration: ' + key)
    git('config', '--local', key, value)
git('pull', '--ff-only')
files = sorted((root / 'docs' / 'wiki').glob('*.md'))
for source in files:
    text = source.read_text()
    text = re.sub(r'\(([A-Za-z_-]+)\.md\)', r'(\1)', text)
    (wiki / source.name).write_text(text)
git('add', '--', *[p.name for p in files])
changed = subprocess.run(['git', 'diff', '--cached', '--quiet'], cwd=wiki).returncode
if changed:
    git('commit', '-S', '-m', 'docs: publish SINet guides and prototype validation')
    git('verify-commit', 'HEAD')
    git('push', 'origin', 'HEAD')
local = git('rev-parse', 'HEAD')
remote_sha = git('ls-remote', 'origin', 'HEAD').split()[0]
if local != remote_sha:
    raise SystemExit('Wiki remote SHA does not match signed local HEAD')
print('Wiki signed commit and exact remote SHA verified:', local)
