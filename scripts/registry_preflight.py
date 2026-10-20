"""Refuse conflicting immutable registry versions; permit identical publish retries."""
import argparse
import base64
import hashlib
import json
import os
from pathlib import Path
import urllib.error
import urllib.request


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--channel', choices=['npm', 'pypi'], required=True)
    parser.add_argument('--packages', type=Path, required=True)
    args = parser.parse_args()
    version = Path('VERSION').read_text().strip()
    if args.channel == 'npm':
        url = 'https://registry.npmjs.org/@ao3575911%2fsinet/' + version
    else:
        url = f'https://pypi.org/pypi/sinet-cli/{version}/json'
    try:
        with urllib.request.urlopen(url, timeout=20) as response:
            metadata = json.load(response)
    except urllib.error.HTTPError as error:
        if error.code != 404:
            raise
        metadata = None
    complete = False
    if metadata is not None:
        if args.channel == 'npm':
            tarball = next(args.packages.glob('*.tgz'))
            expected = 'sha512-' + base64.b64encode(hashlib.sha512(tarball.read_bytes()).digest()).decode()
            if metadata['dist']['integrity'] != expected:
                raise SystemExit('Refusing to replace a conflicting npm version')
            complete = True
        else:
            expected = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in args.packages.glob('*.whl')}
            actual = {item['filename']: item['digests']['sha256'] for item in metadata['urls']}
            if any(expected.get(name) != checksum for name, checksum in actual.items()):
                raise SystemExit('Refusing to add files to a conflicting PyPI version')
            complete = actual == expected
    state = str(complete).lower()
    print('Registry version already complete:', state)
    if os.environ.get('GITHUB_OUTPUT'):
        with open(os.environ['GITHUB_OUTPUT'], 'a') as output:
            output.write('published=' + state + '\n')


if __name__ == '__main__':
    main()
