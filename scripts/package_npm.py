"""Prepare an npm package from pinned, previously released CLI archives."""
import hashlib
import json
from pathlib import Path
import shutil
import tarfile
import zipfile

ROOT = Path(__file__).resolve().parent.parent
ARCHIVES = {
    'linux-amd64': ('tar.gz', 'a3d1de8c480abe462389df94fbb51f2e750e3702ea854d492a287301fe62cdbb'),
    'darwin-arm64': ('tar.gz', '9f92b5c5a393f83904454c31864a9f6b75df5a4d17ce8a7e95bbc91f28bcd917'),
    'windows-amd64': ('zip', '0a26099e1e65f51201af7a1e8a85320bfcc656d7dcfd690a05a1c3649738a35f'),
}


def main():
    package = json.loads((ROOT / 'npm/package.json').read_text())
    if package['version'] != '0.1.1' or (ROOT / 'VERSION').read_text().strip() != '0.1.1':
        raise SystemExit('Update the pinned release archives before packaging another version')
    source = ROOT / 'work/npm-release'
    output = ROOT / 'work/npm-package'
    if output.exists():
        raise SystemExit('work/npm-package must not exist; preserve previous packages separately')
    # Verify every input before creating any output.
    for platform, (extension, digest) in ARCHIVES.items():
        archive = source / f'SINet-0.1.1-{platform}.{extension}'
        if hashlib.sha256(archive.read_bytes()).hexdigest() != digest:
            raise SystemExit(f'Checksum mismatch: {archive.name}')
    shutil.copytree(ROOT / 'npm', output)
    for name in ('LICENSE', 'THIRD_PARTY_NOTICES.md'):
        shutil.copyfile(ROOT / name, output / name)
    for platform, (extension, _) in ARCHIVES.items():
        stem = f'SINet-0.1.1-{platform}'
        binary = 'SINet.exe' if platform == 'windows-amd64' else 'SINet'
        archive = source / f'{stem}.{extension}'
        if extension == 'zip':
            with zipfile.ZipFile(archive) as bundle:
                content = bundle.read(f'{stem}/{binary}')
        else:
            with tarfile.open(archive) as bundle:
                content = bundle.extractfile(f'{stem}/{binary}').read()
        target = output / 'binaries' / platform / binary
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(content)
        target.chmod(0o755)
    (output / 'cli.cjs').chmod(0o755)
    print(output)


if __name__ == '__main__':
    main()
