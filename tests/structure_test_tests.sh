#!/bin/bash

# Copyright 2017 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


#Structure test tests. The tests to test the structure tests.

#End to end tests to make sure the structure tests do what we expect them
#to do on a known quantity, the latest debian docker image.

failures=0
# build newest structure test binary
make cross
make image

test_dir=$(dirname "$0")
# Run the debian tests, they should always pass on latest
test_image="gcr.io/google-appengine/debian8"
docker pull "$test_image"

res=$(./out/container-structure-test test --image "$test_image" --config "$test_dir"/debian_test.yaml)
code=$?

if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "Success case test failed"
  echo "$res"
  failures=$((failures +1))
fi

# test image metadata
run_metadata_tests=true
if $run_metadata_tests ;
then
  test_metadata_image=debian8-with-metadata:latest
  docker build -q -f "$test_dir"/Dockerfile.metadata --tag "$test_metadata_image" "$test_dir"
  res=$(./out/container-structure-test test --image "$test_metadata_image" --config "$test_dir"/debian_metadata_test.yaml)
  code=$?

  if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
  then
    echo "Metadata success case test failed"
    echo "$res"
    failures=$((failures +1))
  fi
  docker rmi "$test_metadata_image"
fi

# Run some bogus tests, they should fail as expected
res=$(./out/container-structure-test test --image "$test_image" --config "$test_dir"/debian_failure_test.yaml)
code=$?

if ! [[ ("$res" =~ "FAIL" && "$code" == "1") ]];
then
  echo "Failure case test failed"
  echo "$res"
  failures=$((failures +1))
fi

# Test the image.
abs_test_dir=$(readlink -f "$test_dir")
res=$(docker run -v /var/run/docker.sock:/var/run/docker.sock \
                 -v "$abs_test_dir":/tests \
                 gcr.io/gcp-runtimes/container-structure-test:latest test --image "$test_image" --config /tests/debian_test.yaml)
code=$?

if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "Image success case test failed"
  echo "$res"
  failures=$((failures +1))
fi

echo "Failure Count: $failures"
if [ "$failures" -gt "0" ]
then
  exit 1
fi
