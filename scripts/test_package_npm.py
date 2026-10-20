"""Release installer integrity gates, without network or registry credentials."""
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import package_npm


class PackageIntegrityTests(unittest.TestCase):
    def prepare(self, directory, version='0.1.1'):
        root = Path(directory)
        (root / 'npm').mkdir()
        (root / 'npm/package.json').write_text(json.dumps({'version': version}))
        (root / 'VERSION').write_text('0.1.1')
        (root / 'work/npm-release').mkdir(parents=True)
        return root

    def test_modified_archive_is_rejected_before_output(self):
        with tempfile.TemporaryDirectory() as directory:
            root = self.prepare(directory)
            (root / 'work/npm-release/SINet-0.1.1-linux-amd64.tar.gz').write_bytes(b'tampered')
            with patch.object(package_npm, 'ROOT', root):
                with self.assertRaisesRegex(SystemExit, 'Checksum mismatch'):
                    package_npm.main()
            self.assertFalse((root / 'work/npm-package').exists())

    def test_new_version_requires_new_release_pins(self):
        with tempfile.TemporaryDirectory() as directory:
            root = self.prepare(directory, '0.1.2')
            with patch.object(package_npm, 'ROOT', root):
                with self.assertRaisesRegex(SystemExit, 'Update the pinned release archives'):
                    package_npm.main()
            self.assertFalse((root / 'work/npm-package').exists())


if __name__ == '__main__':
    unittest.main()
