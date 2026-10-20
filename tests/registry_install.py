"""Exercise real npm/npx/pip installs of local candidates, including tamper rejection."""
import argparse
import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parent.parent


def run(args, env=None):
    return subprocess.check_output(list(map(str, args)), text=True, env=env).strip()


def version(args, expected, env=None):
    result = json.loads(run(args, env=env))
    assert result['name'] == 'SINet' and result['version'] == expected, result


def tamper(binary, command, env=None):
    content = bytearray(binary.read_bytes())
    content[0] ^= 1
    binary.write_bytes(content)
    result = subprocess.run(list(map(str, command)), text=True, capture_output=True, env=env)
    assert result.returncode != 0 and 'integrity' in result.stderr, result.stderr


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--packages', type=Path, required=True)
    args = parser.parse_args()
    packages = args.packages.resolve()
    expected = (ROOT / 'VERSION').read_text().strip()
    npm = shutil.which('npm')
    npx = shutil.which('npx')
    node = shutil.which('node')
    assert npm and npx and node, 'Node/npm/npx required'
    targets = {('linux', 'x86_64'): 'SINet-linux-amd64',
               ('darwin', 'arm64'): 'SINet-darwin-arm64',
               ('win32', 'amd64'): 'SINet-windows-amd64.exe'}
    filename = targets[(sys.platform, platform.machine().lower())]
    with tempfile.TemporaryDirectory(prefix='registry-install-') as directory:
        temp = Path(directory)
        env = dict(os.environ, npm_config_cache=str(temp / 'npm-cache'), npm_config_ignore_scripts='true')
        tarball = next(packages.glob('*.tgz'))
        prefix = temp / 'npm-prefix'
        run([npm, 'install', '--global', '--ignore-scripts', '--prefix', prefix, tarball], env=env)
        module = prefix / ('node_modules' if os.name == 'nt' else 'lib/node_modules') / '@ao3575911/sinet'
        version([node, module / 'bin/cli.cjs', 'version'], expected, env)
        version([npx, '--yes', '--package', tarball, 'SINet', '--version'], expected, env)
        tamper(module / 'bin' / filename, [node, module / 'bin/cli.cjs', 'version'], env)
        subprocess.run([sys.executable, '-m', 'venv', str(temp / 'venv')], check=True)
        bin_dir = temp / 'venv' / ('Scripts' if os.name == 'nt' else 'bin')
        python = bin_dir / ('python.exe' if os.name == 'nt' else 'python')
        subprocess.run([str(python), '-m', 'pip', 'install', '--no-index', '--no-deps',
                        '--find-links', str(packages), 'sinet-cli==' + expected], check=True)
        cli = bin_dir / ('SINet.exe' if os.name == 'nt' else 'SINet')
        version([cli, 'version'], expected)
        installed = Path(run([python, '-c', 'import sinet_cli; print(sinet_cli.__file__)'])).parent
        tamper(installed / 'bin' / filename, [cli, 'version'])
    print('PASS: npm install, npx, pip install, both launchers and both tamper denials')


if __name__ == '__main__':
    main()
