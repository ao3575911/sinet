"""Run the bundled SINet Go CLI; no network or installation hooks."""
import hashlib
import json
import os
from pathlib import Path
import platform
import stat
import subprocess
import sys


def main():
    targets = {('linux', 'x86_64'): 'SINet-linux-amd64',
               ('darwin', 'arm64'): 'SINet-darwin-arm64',
               ('win32', 'amd64'): 'SINet-windows-amd64.exe'}
    try:
        filename = targets.get((sys.platform, platform.machine().lower()))
        if filename is None:
            raise ValueError('Supported targets: Linux x64, macOS arm64, Windows x64.')
        directory = Path(__file__).parent / 'bin'
        binary = directory / filename
        manifest = json.loads((directory / 'SHA256.json').read_text())
        if not stat.S_ISREG(binary.lstat().st_mode):
            raise ValueError('SINet binary must be a regular file.')
        expected = manifest.get(filename, '')
        if len(expected) != 64 or hashlib.sha256(binary.read_bytes()).hexdigest() != expected:
            raise ValueError('SINet binary integrity check failed; reinstall the package.')
        if os.name != 'nt':
            os.execv(str(binary), [str(binary), *sys.argv[1:]])
        return subprocess.run([str(binary), *sys.argv[1:]], check=False).returncode
    except KeyboardInterrupt:
        return 130
    except (OSError, ValueError, KeyError) as error:
        print(f'SINet: {error}', file=sys.stderr)
        return 1
