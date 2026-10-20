#!/usr/bin/env python3
"""Validate SemVer 2.0.0, changelog, optional tag and compiled CLI agreement."""
import argparse
import json
import pathlib
import re
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[1]
NUMBER = r'(?:0|[1-9][0-9]*)'
PRERELEASE = r'(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)'
PATTERN = re.compile(rf'{NUMBER}\.{NUMBER}\.{NUMBER}(?:-{PRERELEASE}(?:\.{PRERELEASE})*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?')


def valid(value):
    return len(value) <= 128 and PATTERN.fullmatch(value) is not None


def check(tag=None, binary=None):
    raw = (ROOT / 'VERSION').read_text()
    version = raw.strip()
    if raw != version + '\n' or not valid(version):
        raise ValueError('VERSION must be a single canonical SemVer 2.0.0 line')
    if f'## [{version}] - ' not in (ROOT / 'CHANGELOG.md').read_text():
        raise ValueError('CHANGELOG.md must contain a dated entry for VERSION')
    if tag is not None and tag != 'v' + version:
        raise ValueError('Release tag must equal v + VERSION')
    if binary:
        output = subprocess.check_output([binary, 'version'], text=True)
        data = json.loads(output)
        if data.get('name') != 'SINet' or data.get('version') != version:
            raise ValueError('Compiled SINet version disagrees with VERSION')
    return version


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--tag')
    parser.add_argument('--binary')
    args = parser.parse_args()
    print('PASS SemVer:', check(args.tag, args.binary))
