import unittest
from check_version import valid, check, ROOT


class VersionTests(unittest.TestCase):
    def test_semver_valid(self):
        for value in ['0.1.0', '1.0.0', '2.1.3-rc.1', '1.2.3-alpha.0+build.001', '0.0.1-x-y.z']:
            with self.subTest(value=value):
                self.assertTrue(valid(value))

    def test_semver_invalid(self):
        for value in ['v0.1.0', '01.2.3', '1.02.3', '1.2', '1.2.3-01', '1.2.3-a..b',
                      '1.2.3+', '1.2.3.4', '1.2.3\n', '1.2.3-α', '1.2.3;echo bad']:
            with self.subTest(value=value):
                self.assertFalse(valid(value))

    def test_tag_binding_and_changelog(self):
        version=(ROOT / 'VERSION').read_text().strip()
        self.assertEqual(check(tag='v'+version), version)
        with self.assertRaises(ValueError):
            check(tag='v'+version+'-unexpected')

    def test_publication_intents_and_retries(self):
        version = (ROOT / 'VERSION').read_text().strip()
        for channel in ['npm', 'pypi']:
            for suffix in ['', '-retry.1', '-retry.12']:
                self.assertEqual(check(publication_tag=channel+'-v'+version+suffix), version)

    def test_publication_intents_reject_wrong_version_or_retry(self):
        version = (ROOT / 'VERSION').read_text().strip()
        for tag in ['npm-v999.0.0', 'other-v'+version, 'npm-v'+version+'-retry.0',
                    'npm-v'+version+'-retry.01', 'npm-v'+version+'-retry.x',
                    'npm-v'+version+'; echo unsafe']:
            with self.subTest(tag=tag), self.assertRaises(ValueError):
                check(publication_tag=tag)


if __name__ == '__main__':
    unittest.main()
