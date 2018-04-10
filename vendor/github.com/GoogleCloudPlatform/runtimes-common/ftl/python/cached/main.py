# Copyright 2017 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import argparse
import sys
from ftl.cached import args
from ftl.cached import cached

_RUNTIME = "python"

parser = args.base_parser()
python_parser = argparse.ArgumentParser(
    add_help=False, parents=[parser], description='Run python benchmark.')


def main(cli_args):
    parsed_args = python_parser.parse_args(cli_args)
    c = cached.Cached(parsed_args, _RUNTIME)
    c.run_cached_tests()


if __name__ == '__main__':
    main(sys.argv[1:])
