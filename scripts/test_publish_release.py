import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import publish_release


class PublisherTests(unittest.TestCase):
    def exercise(self, digest):
        with tempfile.TemporaryDirectory() as directory:
            previous = os.getcwd()
            os.chdir(directory)
            try:
                Path('VERSION').write_text('0.1.1\n')
                Path('dist').mkdir()
                Path('dist/package.zip').write_bytes(b'package')
                release = {'tag_name': 'v0.1.1', 'html_url': 'https://example.test/release',
                           'assets': [{'name': 'package.zip', 'digest': digest}]}
                calls = []

                def gh(*args):
                    calls.append(args)
                    if args == ('api', 'repos/ao3575911/sinet/releases?per_page=100'):
                        return json.dumps([release])
                    if args[:2] == ('release', 'edit'):
                        return ''
                    raise AssertionError(f'Unexpected API operation: {args}')

                result = subprocess.CompletedProcess([], 0, '{"isDraft":true}', '')
                with patch('sys.argv', ['publish_release.py', '--existing-only']), \
                     patch.object(publish_release.subprocess, 'run', return_value=result), \
                     patch.object(publish_release, 'gh', side_effect=gh):
                    publish_release.main()
                return calls
            finally:
                os.chdir(previous)

    def test_draft_uses_authenticated_listing_and_verifies_digest(self):
        calls = self.exercise('sha256:' + hashlib.sha256(b'package').hexdigest())
        self.assertTrue(any(c[:2] == ('release', 'edit') for c in calls))
        self.assertFalse(any(c[:2] == ('release', 'upload') for c in calls))

    def test_mismatch_prevents_publication(self):
        with self.assertRaisesRegex(SystemExit, 'digest mismatch'):
            self.exercise('sha256:incorrect')


if __name__ == '__main__':
    unittest.main()
