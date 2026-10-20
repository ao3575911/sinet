"""Publish only after attestation verification; never replace published assets."""
import hashlib
import json
from pathlib import Path
import subprocess


def gh(*args):
    return subprocess.check_output(['gh', *args], text=True)


def main():
    tag = 'v' + Path('VERSION').read_text().strip()
    repo = 'ao3575911/sinet'
    assets = sorted(Path('dist').iterdir())
    result = subprocess.run(['gh', 'release', 'view', tag, '--repo', repo,
                             '--json', 'isDraft'], capture_output=True, text=True)
    if result.returncode:
        gh('release', 'create', tag, '--repo', repo, '--verify-tag', '--draft',
           '--prerelease', '--title', f'SINet {tag} — prototype',
           '--notes-file', 'docs/RELEASE.md')
        draft = True
    else:
        draft = json.loads(result.stdout)['isDraft']
    if draft:
        gh('release', 'upload', tag, '--repo', repo, '--clobber', *map(str, assets))
    release = json.loads(gh('api', f'repos/{repo}/releases/tags/{tag}'))
    remote = {item['name']: item for item in release['assets']}
    if set(remote) != {p.name for p in assets}:
        raise SystemExit('Unexpected release asset set')
    for path in assets:
        expected = 'sha256:' + hashlib.sha256(path.read_bytes()).hexdigest()
        if remote[path.name].get('digest') != expected:
            raise SystemExit(f'Remote digest mismatch: {path.name}')
    if draft:
        gh('release', 'edit', tag, '--repo', repo, '--draft=false', '--prerelease')
    print(f'Published and verified: {release["html_url"]}')


if __name__ == '__main__':
    main()
