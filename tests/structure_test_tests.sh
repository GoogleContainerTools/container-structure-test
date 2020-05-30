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

# Get the architecture to load the right configurations
go_architecture=$(go env GOARCH)

# Get the absolute path of the tests directory
test_dir="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
test_config_dir="${test_dir}/${go_architecture}"

# If a configuration folder for the architecture doesn't exist, default to amd64
test -d ${test_config_dir} || test_config_dir="${test_dir}/amd64"

echo "##"
echo "# Build the newest 'container structure test' binary"
echo "##"
cp -f "${test_dir}/Dockerfile" "${test_dir}/../Dockerfile"
make
make cross
make image

# Run the ubuntu tests, they should always pass on 20.04
test_image="ubuntu:20.04"
docker pull "$test_image" > /dev/null

echo "##"
echo "# Positive Test Case"
echo "##"
res=$(./out/container-structure-test test --image "$test_image" --config "${test_config_dir}/ubuntu_20_04_test.yaml")
code=$?
if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "FAIL: Success test case failed"
  echo "$res"
  failures=$((failures +1))
else
  echo "PASS: Success test case passed"
fi

echo "##"
echo "# Metadata Test Case"
echo "##"
# test image metadata
run_metadata_tests=true
if $run_metadata_tests ;
then
  test_metadata_image=debian8-with-metadata:latest
  test_metadata_tar=debian8-with-metadata.tar
  test_metadata_dir=debian8-with-metadata
  docker build -q -f "$test_dir"/Dockerfile.metadata --tag "$test_metadata_image" "$test_dir" > /dev/null
  res=$(./out/container-structure-test test --image "$test_metadata_image" --config "${test_config_dir}/ubuntu_20_04_metadata_test.yaml")
  code=$?

  if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
  then
    echo "Metadata success case test failed for docker driver"
    echo "$res"
    failures=$((failures +1))
  fi

  docker save "$test_metadata_image" -o "$test_metadata_tar" > /dev/null
  res=$(./out/container-structure-test test --driver tar --image "$test_metadata_tar" --config "${test_config_dir}/ubuntu_20_04_metadata_test.yaml")
  code=$?
  if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
  then
    echo "Metadata success case test failed for tar driver"
    echo "$res"
    failures=$((failures +1))
  fi

  mkdir -p "$test_metadata_dir"
  tar -C "$test_metadata_dir" -xvf "$test_metadata_tar" > /dev/null
  test_metadata_json=$(grep 'Config":"\K[^"]+' -Po "$test_metadata_dir/manifest.json")
  res=$(./out/container-structure-test test --driver host --force --metadata "$test_metadata_dir/$test_metadata_json" --config "${test_config_dir}/ubuntu_20_04_metadata_test.yaml")
  code=$?
  if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
  then
    echo "Metadata success case test failed for host driver"
    echo "$res"
    failures=$((failures +1))
  fi

  rm -rf "$test_metadata_dir"
  rm "$test_metadata_tar"
  docker rmi "$test_metadata_image" > /dev/null
fi

echo "##"
echo "# Failure Test Case"
echo "##"
# Run some bogus tests, they should fail as expected
res=$(./out/container-structure-test test --image "$test_image" --config "${test_config_dir}/ubuntu_20_04_failure_test.yaml")
code=$?
if ! [[ ("$res" =~ "FAIL" && "$code" == "1") ]];
then
  echo "FAIL: Failure test case did not fail"
  echo "$res"
  failures=$((failures +1))
else
  echo "PASS: Failure test failed"
fi

# Test the image.
res=$(docker run -v /var/run/docker.sock:/var/run/docker.sock \
                 -v "$test_config_dir":/tests \
                 gcr.io/gcp-runtimes/container-structure-test:latest test --image "$test_image" --config "/tests/ubuntu_20_04_test.yaml")
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
