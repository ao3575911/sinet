"""Build portable prototype packages using Go and Python's standard library."""
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import subprocess
import tarfile
import zipfile


def run(*args, **kwargs):
    return subprocess.check_output(args, text=True, **kwargs).strip()


def main():
    version = Path('VERSION').read_text().strip()
    run('python3', 'scripts/check_version.py')
    output = Path('dist')
    output.mkdir(exist_ok=True)
    if list(output.iterdir()):
        raise SystemExit('dist must be empty; keep previous release artifacts separately')
    build = Path('work/release-build')
    build.mkdir(parents=True, exist_ok=True)
    for component, target, arch in [('SINet', 'linux', 'amd64'),
                                    ('SINet', 'darwin', 'arm64'),
                                    ('SINet', 'windows', 'amd64'),
                                    ('server', 'linux', 'amd64')]:
        binary = component + ('.exe' if target == 'windows' else '')
        destination = build / binary
        env = dict(os.environ, GOOS=target, GOARCH=arch, CGO_ENABLED='0')
        subprocess.run(['go', 'build', '-trimpath', '-buildvcs=false',
                        '-ldflags=-s -w', '-o', str(destination), './cmd/' + component],
                       env=env, check=True)
        root = f'{component}-{version}-{target}-{arch}'
        files = {binary: destination.read_bytes(),
                 'LICENSE': Path('LICENSE').read_bytes(),
                 'THIRD_PARTY_NOTICES.md': Path('THIRD_PARTY_NOTICES.md').read_bytes(),
                 'CHANGELOG.md': Path('CHANGELOG.md').read_bytes(),
                 'INSTALL.md': Path('docs/RELEASE.md').read_bytes()}
        if component == 'server':
            files['.env.example'] = Path('.env.example').read_bytes()
            files['OPERATIONS.md'] = Path('docs/OPERATIONS.md').read_bytes()
        if target == 'windows':
            with zipfile.ZipFile(output / (root + '.zip'), 'w', zipfile.ZIP_DEFLATED) as archive:
                for name, content in sorted(files.items()):
                    entry = zipfile.ZipInfo(root + '/' + name, (1980, 1, 1, 0, 0, 0))
                    entry.external_attr = (0o755 if name == binary else 0o644) << 16
                    archive.writestr(entry, content, compress_type=zipfile.ZIP_DEFLATED)
        else:
            with (output / (root + '.tar.gz')).open('wb') as raw:
                with gzip.GzipFile(fileobj=raw, mode='wb', filename='', mtime=0) as compressed:
                    with tarfile.open(fileobj=compressed, mode='w') as archive:
                        for name, content in sorted(files.items()):
                            entry = tarfile.TarInfo(root + '/' + name)
                            entry.size = len(content)
                            entry.mode = 0o755 if name == binary else 0o644
                            archive.addfile(entry, io.BytesIO(content))
    info = {'version': version, 'source_commit': run('git', 'rev-parse', 'HEAD'),
            'compiler': run('go', 'version'), 'status': 'prototype',
            'cgo': False, 'build_flags': ['-trimpath', '-buildvcs=false', '-ldflags=-s -w']}
    (output / 'BUILDINFO.json').write_text(json.dumps(info, indent=2) + '\n')
    sums = ''.join(f'{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.name}\n'
                   for p in sorted(output.iterdir()))
    (output / 'SHA256SUMS').write_text(sums)


if __name__ == '__main__':
    main()
