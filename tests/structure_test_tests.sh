#!/usr/bin/env bash

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
test_dir="$( cd "$(dirname "$0")" || return >/dev/null 2>&1 ; pwd -P )"
test_config_dir="${test_dir}/${go_architecture}"

# If a configuration folder for the architecture doesn't exist, default to amd64
test -d ${test_config_dir} || test_config_dir="${test_dir}/amd64"


function HEADER() {
  local msg="$1"
  echo ""
  echo "###############"
  echo "# $msg"
  echo "###############"
  echo ""
}

function build_image() {
  _file=$1
  _tag=$2
  _dir=$3
  docker build -q -f "$_dir/$_file" --tag "$_tag" "$_dir" > /dev/null
}

HEADER "Determine the runtime"

DOCKER=""
if which docker > /dev/null; then
  DOCKER=$(which docker)
  echo "Using docker at $(which docker)"
elif which podman > /dev/null; then 
  DOCKER=$(which podman)
  echo "Using podman at $(which podman)"
else 
  echo "Could not find a runtime to run tests"
  exit 1
fi

docker() {
  $DOCKER "$@"
}


HEADER "Build the newest 'container structure test' binary"

cp -f "${test_dir}/Dockerfile" "${test_dir}/../Dockerfile"
make DOCKER=$DOCKER
make DOCKER=$DOCKER cross
make DOCKER=$DOCKER image

# Run the ubuntu tests, they should always pass on 22.04
test_image="ubuntu:22.04"
docker pull "$test_image" > /dev/null


HEADER "Positive Test Case"

res=$(./out/container-structure-test test --image "$test_image" --config "${test_config_dir}/ubuntu_22_04_test.yaml")
code=$?
if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "FAIL: Success test case failed"
  echo "$res"
  failures=$((failures +1))
else
  echo "PASS: Success test case passed"
fi


HEADER "Container Run Options Test Cases"
test_containeropts_user_image="test.local/ubuntu-unprivileged:latest"
build_image "Dockerfile.unprivileged" "$test_containeropts_user_image" "$test_dir"
res=$(./out/container-structure-test test --image "$test_containeropts_user_image" --config "${test_config_dir}/ubuntu_22_04_containeropts_user_test.yaml")
code=$?
if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "FAIL: Run options (user) test case failed"
  echo "$res"
  failures=$((failures +1))
else
  echo "PASS: Run option (user) test case passed"
fi
docker rmi "$test_containeropts_user_image" > /dev/null

test_containeropts_cap_image="test.local/ubuntu-cap:latest"
build_image "Dockerfile.cap" "$test_containeropts_cap_image" "$test_dir"
res=$(./out/container-structure-test test --image "$test_containeropts_cap_image" --config "${test_config_dir}/ubuntu_22_04_containeropts_test.yaml")
code=$?
if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "FAIL: Run options (capabilities, bindMounts) test case failed"
  echo "$res"
  failures=$((failures +1))
else
  echo "PASS: Run options (capabilities, bindMounts) test case passed"
fi
docker rmi "$test_containeropts_cap_image" > /dev/null

res=$(FOO='keepitsecret!' BAR='keepitsafe!' ./out/container-structure-test test --image "$test_image" --config "${test_config_dir}/ubuntu_22_04_containeropts_env_test.yaml")
code=$?
if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "FAIL: Run options (envVars) test case failed"
  echo "$res"
  failures=$((failures +1))
else
  echo "PASS: Run options (envVars) test case passed"
fi

res=$(./out/container-structure-test test --image "$test_image" --config "${test_config_dir}/ubuntu_22_04_containeropts_envfile_test.yaml")
code=$?
if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "FAIL: Run options (envFile) test case failed"
  echo "$res"
  failures=$((failures +1))
else
  echo "PASS: Run options (envFile) test case passed"
fi


HEADER "Metadata Test Case"
# test image metadata
run_metadata_tests=true
if $run_metadata_tests ;
then
  test_metadata_image=test.local/debian8-with-metadata:latest
  test_metadata_tar=debian8-with-metadata.tar
  test_metadata_dir=debian8-with-metadata
  build_image "Dockerfile.metadata" "$test_metadata_image" "$test_dir"
  res=$(./out/container-structure-test test --image "$test_metadata_image" --config "${test_config_dir}/ubuntu_22_04_metadata_test.yaml")
  code=$?

  if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
  then
    echo "FAIL: Metadata success test case for docker driver"
    echo "$res"
    failures=$((failures +1))
  else
    echo "PASS: Metadata success test case for docker driver"
  fi

  docker save "$test_metadata_image" -o "$test_metadata_tar" > /dev/null
  res=$(./out/container-structure-test test --driver tar --image "$test_metadata_tar" --config "${test_config_dir}/ubuntu_22_04_metadata_test.yaml")
  code=$?
  if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
  then
    echo "FAIL: Metadata success test case for tar driver"
    echo "$res"
    failures=$((failures +1))
  else
    echo "PASS: Metadata success test case for tar driver"
  fi

  mkdir -p "$test_metadata_dir"
  tar -C "$test_metadata_dir" -xf "$test_metadata_tar" > /dev/null
  test_metadata_json=$(jq -r '.[0].Config' "$test_metadata_dir/manifest.json")
  res=$(./out/container-structure-test test --driver host --force --metadata "$test_metadata_dir/$test_metadata_json" --config "${test_config_dir}/ubuntu_22_04_metadata_test.yaml")
  code=$?
  if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
  then
    echo "FAIL: Metadata success test case for host driver"
    echo "$res"
    failures=$((failures +1))
  else
    echo "PASS: Metadata success test case for host driver"
  fi

  rm -rf "$test_metadata_dir"
  rm "$test_metadata_tar"
  docker rmi "$test_metadata_image" > /dev/null
fi


HEADER "Failure Test Case"

# Run some bogus tests, they should fail as expected
res=$(./out/container-structure-test test --image "$test_image" --config "${test_config_dir}/ubuntu_22_04_failure_test.yaml")
code=$?
if ! [[ ("$res" =~ "FAIL" && "$code" == "1") ]];
then
  echo "FAIL: Failure test case did not fail"
  echo "$res"
  failures=$((failures +1))
else
  echo "PASS: Failure test failed"
fi



HEADER "OCI layout test case"

go install github.com/google/go-containerregistry/cmd/crane
tmp="$(mktemp -d)"

crane pull "$test_image" --format=oci "$tmp" --platform="linux/$go_architecture"


res=$(./out/container-structure-test test --image-from-oci-layout="$tmp" --config "${test_config_dir}/ubuntu_22_04_test.yaml" 2>&1)
code=$?
if ! [[ ("$res" =~ index\ does\ not\ contain\ a\ reference\ annotation\.\ \-\-default\-image\-tag\ must\ be\ provided\. && "$code" == "1") ]];
then
  echo "FAIL: oci failing test case"
  echo "$res"
  echo "$code"
  failures=$((failures +1))
else
  echo "PASS: oci failing test case"
fi

res=$(./out/container-structure-test test --image-from-oci-layout="$tmp" --default-image-tag="test.local/$test_image" --config "${test_config_dir}/ubuntu_22_04_test.yaml" 2>&1)
code=$?
if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "FAIL: oci success test case"
  echo "$res"
  echo "$code"
  failures=$((failures +1))
else
  echo "PASS: oci success test case"
fi

HEADER "Platform test cases"

docker run --rm --privileged tonistiigi/binfmt --install all > /dev/null
res=$(./out/container-structure-test test --image "$test_image" --platform="linux/$go_architecture" --config "${test_config_dir}/ubuntu_22_04_test.yaml" 2>&1)
code=$?
if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "FAIL: current host platform test case"
  echo "$res"
  echo "$code"
  failures=$((failures +1))
else
  echo "PASS: current host platform test case"
fi

res=$(./out/container-structure-test test --image "$test_image" --platform="linux/riscv64" --config "${test_config_dir}/ubuntu_22_04_test.yaml" 2>&1)
code=$?
if ! [[ ("$res" =~ image\ with\ reference.+was\ found\ but\ does\ not\ match\ the\ specified\ platform:\ wanted\ linux\/\riscv64,\ actual:\ linux\/$go_architecture && "$code" == "1") ]];
then
  echo "FAIL: platform failing test case"
  echo "$res"
  echo "$code"
  failures=$((failures +1))
else
  echo "PASS: platform failing test case"
fi

test_config_dir="${test_dir}/s390x"
res=$(./out/container-structure-test test --image "$test_image" --platform="linux/s390x" --pull --config "${test_config_dir}/ubuntu_22_04_test.yaml" 2>&1)
code=$?
if ! [[ ("$res" =~ "PASS" && "$code" == "0") ]];
then
  echo "FAIL: platform w/ --pull test case"
  echo "$res"
  echo "$code"
  failures=$((failures +1))
else
  echo "PASS: platform w/ --pull test case"
fi


if [ $failures -gt 0 ]; then
  echo "Some tests did not pass. $failures"
  exit 1
fi
