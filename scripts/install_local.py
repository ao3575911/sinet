#!/usr/bin/env python3
"""Install a reviewed local build; no downloads or dependency on the current directory."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parent.parent


def replace_link(path, target):
    path.parent.mkdir(parents=True, exist_ok=True)
    if path.is_symlink() and path.resolve() == target.resolve():
        return
    if path.exists() or path.is_symlink():
        backup = path.with_name(path.name + f'.before-sinet-{time.time_ns()}')
        # Back up without removing the working launcher until replacement is ready.
        if path.is_symlink():
            backup.symlink_to(os.readlink(path))
        else:
            shutil.copy2(path, backup)
        print(f'Previous launcher saved: {backup}')
    with tempfile.TemporaryDirectory(prefix='.sinet-link-', dir=path.parent) as tmp:
        link = Path(tmp) / 'SINet'
        link.symlink_to(target)
        os.replace(link, path)


def install(binary, prefix, entry_points):
    binary = binary.resolve(strict=True)
    expected = (ROOT / 'VERSION').read_text().strip()
    reported = json.loads(subprocess.check_output([str(binary), 'version'], timeout=10))
    if reported != {'name': 'SINet', 'version': expected}:
        raise ValueError('binary version does not match VERSION; run make build')
    canonical = prefix / 'share' / 'SINet' / 'bin' / 'SINet'
    canonical.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='.sinet-install-', dir=canonical.parent) as tmp:
        staged = Path(tmp) / 'SINet'
        shutil.copyfile(binary, staged)
        staged.chmod(0o755)
        os.replace(staged, canonical)
    for path in [prefix / 'bin' / 'SINet', *entry_points]:
        # abspath keeps an existing symlink as the entry point, not its target.
        replace_link(Path(os.path.abspath(path)), canonical)
    print(f'Installed SINet {expected}: {canonical}')
    print('Run hash -r in existing shells, then SINet doctor. Pinned package versions stay unchanged.')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, default=ROOT / 'bin' / 'SINet')
    parser.add_argument('--prefix', type=Path, default=Path.home() / '.local')
    parser.add_argument('--entry-point', type=Path, action='append', default=[],
                        help='additional launcher to back up and point at the same binary')
    args = parser.parse_args()
    try:
        install(args.binary, args.prefix.absolute(), args.entry_point)
    except (OSError, ValueError, subprocess.SubprocessError) as exc:
        parser.exit(1, f'Install failed: {exc}\n')


if __name__ == '__main__':
    main()
