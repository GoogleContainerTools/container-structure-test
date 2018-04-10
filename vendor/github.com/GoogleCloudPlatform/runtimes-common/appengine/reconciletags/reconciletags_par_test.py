"""Tests for reconciletags_par."""

import subprocess
import unittest


class ReconciletagsParTest(unittest.TestCase):
    _FILE_NAME = 'retagger.json'

    # Just make sure the par file runs on a real image without failing.
    def testParFile(self):
        subprocess.call(['./appengine/reconciletags/reconciletags.par',
                         '--dry-run',
                         'appengine/reconciletags/'+self._FILE_NAME])


if __name__ == '__main__':
    unittest.main()
