import hashlib
import io
import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import urllib.error
import base64

import registry_preflight


class RegistryPreflightTests(unittest.TestCase):
    def check(self, metadata=None, error=None, channel='pypi'):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'candidate.whl').write_bytes(b'candidate')
            (root / 'candidate.tgz').write_bytes(b'candidate')
            output = root / 'output'
            with patch('sys.argv', ['registry_preflight.py', '--channel', channel, '--packages', str(root)]), \
                 patch.dict(os.environ, {'GITHUB_OUTPUT': str(output)}), \
                 patch.object(registry_preflight.urllib.request, 'urlopen',
                              side_effect=error,
                              return_value=io.BytesIO(json.dumps(metadata).encode())):
                registry_preflight.main()
            return output.read_text()

    def test_identical_existing_version_is_complete(self):
        metadata = {'urls': [{'filename': 'candidate.whl',
                             'digests': {'sha256': hashlib.sha256(b'candidate').hexdigest()}}]}
        self.assertEqual(self.check(metadata), 'published=true\n')

    def test_conflicting_version_fails_closed(self):
        with self.assertRaisesRegex(SystemExit, 'conflicting PyPI version'):
            self.check({'urls': [{'filename': 'candidate.whl', 'digests': {'sha256': 'wrong'}}]})

    def test_missing_version_is_publishable(self):
        error = urllib.error.HTTPError('https://pypi.org', 404, 'missing', {}, None)
        self.assertEqual(self.check(error=error), 'published=false\n')

    def test_auth_or_network_denial_is_not_missing_version(self):
        error = urllib.error.HTTPError('https://pypi.org', 403, 'denied', {}, None)
        with self.assertRaises(urllib.error.HTTPError):
            self.check(error=error)

    def test_identical_npm_integrity_is_complete(self):
        expected = 'sha512-' + base64.b64encode(hashlib.sha512(b'candidate').digest()).decode()
        self.assertEqual(self.check({'dist': {'integrity': expected}}, channel='npm'), 'published=true\n')

    def test_conflicting_npm_version_fails_closed(self):
        with self.assertRaisesRegex(SystemExit, 'conflicting npm version'):
            self.check({'dist': {'integrity': 'wrong'}}, channel='npm')


if __name__ == '__main__':
    unittest.main()
