"""Package previously attested Go release bytes for npm and PyPI, without install hooks."""
import argparse
import base64
import csv
import hashlib
import io
import json
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parent.parent
TARGETS = [
    ('linux-amd64', 'SINet', 'tar.gz', 'manylinux_2_28_x86_64', b'\x7fELF'),
    ('darwin-arm64', 'SINet', 'tar.gz', 'macosx_14_0_arm64', b'\xcf\xfa\xed\xfe'),
    ('windows-amd64', 'SINet.exe', 'zip', 'win_amd64', b'MZ'),
]


def digest(data):
    return hashlib.sha256(data).hexdigest()


def verify_source(source, version):
    names = {'BUILDINFO.json', 'server-' + version + '-linux-amd64.tar.gz'}
    names.update(f'SINet-{version}-{target}.{extension}' for target, _, extension, _, _ in TARGETS)
    checksums = {}
    for line in (source / 'SHA256SUMS').read_text().splitlines():
        match = re.fullmatch(r'([0-9a-f]{64})  ([A-Za-z0-9_.+-]+)', line)
        if not match or match[2] in checksums:
            raise ValueError('Invalid or duplicate release checksum entry')
        checksums[match[2]] = match[1]
    if set(checksums) != names:
        raise ValueError('Unexpected release asset set')
    for name, expected in checksums.items():
        if digest((source / name).read_bytes()) != expected:
            raise ValueError('Release checksum mismatch: ' + name)
    info = json.loads((source / 'BUILDINFO.json').read_text())
    commit = subprocess.check_output(['git', 'rev-parse', f'v{version}^{{commit}}'],
                                     cwd=ROOT, text=True).strip()
    if info['version'] != version or info['source_commit'] != commit or info['cgo'] is not False:
        raise ValueError('Release version, source commit or static build mismatch')


def read_binary(source, version, target, binary, extension, magic):
    member_name = f'SINet-{version}-{target}/{binary}'
    archive_path = source / f'SINet-{version}-{target}.{extension}'
    if extension == 'zip':
        with zipfile.ZipFile(archive_path) as archive:
            member = archive.getinfo(member_name)
            if member.file_size > 32 * 1024 * 1024 or (member.external_attr >> 16) & 0o170000 == 0o120000:
                raise ValueError('Invalid binary archive member')
            content = archive.read(member)
    else:
        with tarfile.open(archive_path) as archive:
            member = archive.getmember(member_name)
            if not member.isreg() or member.size > 32 * 1024 * 1024:
                raise ValueError('Invalid binary archive member')
            content = archive.extractfile(member).read()
    if not content.startswith(magic):
        raise ValueError('Target executable format mismatch')
    return content


def wheel(output, version, target, platform_tag, binary, content, common):
    filename = f'SINet-{target}' + ('.exe' if binary.endswith('.exe') else '')
    distinfo = f'sinet_cli-{version}.dist-info'
    metadata = f'''Metadata-Version: 2.3
Name: sinet-cli
Version: {version}
Summary: The SINet Go social CLI, packaged with a verified binary
Requires-Python: >=3.9
License: Apache-2.0
Project-URL: Source, https://github.com/ao3575911/sinet
Project-URL: Issues, https://github.com/ao3575911/sinet/issues
Description-Content-Type: text/markdown

'''.encode() + common['README.md']
    files = {
        'sinet_cli/__init__.py': (ROOT / 'packaging/pypi/__init__.py').read_bytes(),
        f'sinet_cli/bin/{filename}': content,
        'sinet_cli/bin/SHA256.json': (json.dumps({filename: digest(content)}, sort_keys=True) + '\n').encode(),
        f'{distinfo}/METADATA': metadata,
        f'{distinfo}/WHEEL': f'Wheel-Version: 1.0\nGenerator: SINet stdlib packager\nRoot-Is-Purelib: false\nTag: py3-none-{platform_tag}\n'.encode(),
        f'{distinfo}/entry_points.txt': b'[console_scripts]\nSINet = sinet_cli:main\n',
    }
    for name in ['LICENSE', 'THIRD_PARTY_NOTICES.md', 'BUILDINFO.json']:
        files[f'{distinfo}/{name}'] = common[name]
    record = io.StringIO(newline='')
    writer = csv.writer(record, lineterminator='\n')
    for name, data in sorted(files.items()):
        encoded = base64.urlsafe_b64encode(hashlib.sha256(data).digest()).rstrip(b'=').decode()
        writer.writerow([name, 'sha256=' + encoded, len(data)])
    writer.writerow([f'{distinfo}/RECORD', '', ''])
    files[f'{distinfo}/RECORD'] = record.getvalue().encode()
    with zipfile.ZipFile(output / f'sinet_cli-{version}-py3-none-{platform_tag}.whl', 'w') as archive:
        for name, data in sorted(files.items()):
            entry = zipfile.ZipInfo(name, (1980, 1, 1, 0, 0, 0))
            entry.external_attr = (0o100755 if name == f'sinet_cli/bin/{filename}' else 0o100644) << 16
            archive.writestr(entry, data, compress_type=zipfile.ZIP_DEFLATED)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--source', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=True)
    if list(output.iterdir()):
        raise ValueError('Output directory must be empty')
    version = (ROOT / 'VERSION').read_text().strip()
    if not re.fullmatch(r'(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)', version):
        raise ValueError('Registry packaging currently requires a final x.y.z version')
    verify_source(args.source, version)
    common = {name: (ROOT / name).read_bytes() for name in ['LICENSE', 'THIRD_PARTY_NOTICES.md']}
    common['README.md'] = (ROOT / 'packaging/INSTALL.md').read_bytes()
    common['BUILDINFO.json'] = (args.source / 'BUILDINFO.json').read_bytes()
    (ROOT / 'work').mkdir(exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='registry-build-', dir=ROOT / 'work') as directory:
        package = Path(directory)
        (package / 'bin').mkdir()
        for name, data in common.items():
            (package / name).write_bytes(data)
        launcher = package / 'bin/cli.cjs'
        launcher.write_bytes((ROOT / 'packaging/npm/cli.cjs').read_bytes())
        launcher.chmod(0o755)
        manifest = {}
        for target, binary, extension, platform_tag, magic in TARGETS:
            content = read_binary(args.source, version, target, binary, extension, magic)
            filename = f'SINet-{target}' + ('.exe' if binary.endswith('.exe') else '')
            destination = package / 'bin' / filename
            destination.write_bytes(content)
            destination.chmod(0o755)
            manifest[filename] = digest(content)
            wheel(output, version, target, platform_tag, binary, content, common)
        (package / 'bin/SHA256.json').write_text(json.dumps(manifest, sort_keys=True, indent=2) + '\n')
        metadata = {'name': '@ao3575911/sinet', 'version': version,
                    'description': 'The SINet Go social CLI with verified bundled binaries',
                    'license': 'Apache-2.0', 'bin': {'SINet': 'bin/cli.cjs'},
                    'engines': {'node': '>=22'}, 'os': ['linux', 'darwin', 'win32'],
                    'cpu': ['x64', 'arm64'], 'repository': {'type': 'git', 'url': 'git+https://github.com/ao3575911/sinet.git'},
                    'homepage': 'https://github.com/ao3575911/sinet',
                    'bugs': {'url': 'https://github.com/ao3575911/sinet/issues'},
                    'files': ['bin/', 'README.md', 'LICENSE', 'THIRD_PARTY_NOTICES.md', 'BUILDINFO.json']}
        (package / 'package.json').write_text(json.dumps(metadata, indent=2) + '\n')
        env = dict(os.environ, npm_config_cache=str(ROOT / 'work/npm-cache'))
        subprocess.run(['npm', 'pack', '--ignore-scripts', '--pack-destination', str(output)],
                       cwd=package, env=env, check=True)
    print('Created one npm tarball and three platform wheels in', output)


if __name__ == '__main__':
    main()
