import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import install_local


@unittest.skipIf(os.name == 'nt', 'POSIX symlink installer')
class InstallTests(unittest.TestCase):
    def test_one_binary_backups_and_cwd_independence(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / 'VERSION').write_text('0.2.0-rc.2\n')
            binary = root / 'built'
            binary.write_text('#!/bin/sh\nprintf \'{"name":"SINet","version":"0.2.0-rc.2"}\\n\'\n')
            binary.chmod(0o755)
            prefix = root / 'local'
            (prefix / 'bin').mkdir(parents=True)
            launcher = prefix / 'bin' / 'SINet'
            launcher.write_text('old wrapper')
            extra = root / 'pipx' / 'SINet'
            extra.parent.mkdir()
            extra.symlink_to(launcher)
            with patch.object(install_local, 'ROOT', root):
                install_local.install(binary, prefix, [extra])
                install_local.install(binary, prefix, [extra])
            self.assertEqual(launcher.resolve(), extra.resolve())
            self.assertEqual(len(list(launcher.parent.glob('SINet.before-sinet-*'))), 1)
            self.assertEqual(next(launcher.parent.glob('SINet.before-sinet-*')).read_text(), 'old wrapper')
            for cwd in (root, prefix / 'bin', extra.parent):
                output = subprocess.check_output([str(extra), 'version'], cwd=cwd)
                self.assertEqual(json.loads(output)['version'], '0.2.0-rc.2')

    def test_version_mismatch_leaves_launcher_intact(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / 'VERSION').write_text('0.2.0-rc.2\n')
            binary = root / 'built'
            binary.touch()
            launcher = root / 'bin' / 'SINet'
            launcher.parent.mkdir()
            launcher.write_text('original')
            with patch.object(install_local, 'ROOT', root), patch.object(install_local.subprocess, 'check_output', return_value=b'{"name":"SINet","version":"0.1.1"}'):
                with self.assertRaisesRegex(ValueError, 'does not match'):
                    install_local.install(binary, root, [])
            self.assertEqual(launcher.read_text(), 'original')
