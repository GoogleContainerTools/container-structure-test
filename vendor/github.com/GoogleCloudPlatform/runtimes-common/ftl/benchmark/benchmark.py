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

import subprocess
import datetime
import time
import os
import logging
from google.cloud import bigquery


class Benchmark():
    def __init__(self, args, runtime):
        self._base = args.base
        self._name = args.name
        self._directory = args.directory
        self._iterations = args.iterations
        self._description = args.description
        self._project = args.project
        self._dataset = args.dataset
        self._table = args.table
        self._runtime = runtime
        self._gen_files = args.gen_files

    def _record_build_times_to_bigquery(self, build_times):
        current_date = datetime.datetime.now()
        logging.info('Retrieving bigquery client')
        client = bigquery.Client(project=self._project)

        dataset_ref = client.dataset(self._dataset)
        table_ref = dataset_ref.table(self._table)
        table = client.get_table(table_ref)

        full_name = "{0}:{1}.{2}".format(self._project, self._dataset,
                                         self._table)

        logging.info("Adding build time data to {0}".format(full_name))
        rows = [(current_date, self._description, bt[0], bt[1])
                for bt in build_times]
        client.create_rows(table, rows)
        logging.info("Finished adding build times to {0}".format(full_name))

    def run_benchmarks(self):
        logging.getLogger().setLevel("NOTSET")
        logging.basicConfig(
            format='%(asctime)s.%(msecs)03d %(levelname)-8s %(message)s',
            datefmt='%Y-%m-%d,%H:%M:%S')
        build_times = []
        for i in range(self._gen_files):
            file_name = os.path.join(self._directory, 'app_file_%d' % i)
            with open(file_name, 'wb') as fout:
                fout.write(os.urandom(1024))
        logging.info('Beginning building {0} images'.format(self._runtime))
        for _ in range(self._iterations):
            try:
                start_time = time.time()

                # For the binary
                builder_path = 'ftl/{0}_builder.par'.format(self._runtime)

                # For container builder
                if not os.path.isfile(builder_path):
                    builder_path = 'bazel-bin/ftl/{0}_builder.par'.format(
                        self._runtime)

                cmd = subprocess.Popen(
                    [
                        builder_path, '--base', self._base, '--name',
                        self._name, '--directory', self._directory,
                        '--no-cache'
                    ],
                    stderr=subprocess.PIPE)
                _, output = cmd.communicate()

                build_time = round(time.time() - start_time, 2)
                build_times.append((build_time, output))
            except OSError:
                raise OSError(
                    """Benchmarking assumes either ftl/{0}_builder.par
                    or bazel-bin/ftl/{0}_builder.par
                    exists""".format(self._runtime))
        logging.info('Beginning recording build times to bigquery')
        self._record_build_times_to_bigquery(build_times)
