#!/bin/sh

# Copyright 2016 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

VERBOSE=0
PULL=1
CMD_STRING=""
ENTRYPOINT="/test/structure_test"
ST_IMAGE="gcr.io/gcp-runtimes/structure_test"
USAGE_STRING="Usage: $0 [-i <image>] [-c <config>] [-v] [-e <entrypoint>] [--no-pull]"

CONFIG_DIR=$(pwd)/.cfg
mkdir -p "$CONFIG_DIR"

command -v docker > /dev/null 2>&1 || { echo "Docker is required to run GCP structure tests, but is not installed on this host."; exit 1; }
command docker ps > /dev/null 2>&1 || { echo "Cannot connect to the Docker daemon!"; exit 1; }

cleanup() {
	rm -rf "$CONFIG_DIR"
}

usage() {
	echo "$USAGE_STRING"
	cleanup
	exit 1
}

helper() {
	echo "$USAGE_STRING"
	echo
	echo "	-i, --image          image to run tests on"
	echo "	-c, --config         path to json/yaml config file"
	echo "	-v                   display verbose testing output"
	echo "	-e, --entrypoint     specify custom docker entrypoint for image"
	echo "	--no-pull            don't pull latest structure test image"
	exit 0
}

while test $# -gt 0; do
	case "$1" in
		--image|-i)
			shift
			if test $# -gt 0; then
				IMAGE_NAME=$1
			else
				usage
			fi
			shift
			;;
		--verbose|-v)
			VERBOSE=1
			shift
			;;
		--no-pull)
			PULL=0
			shift
			;;
		--help|-h)
			helper
			;;
		--config|-c)
			shift
			if test $# -eq 0; then
				usage
			else
				if [ ! -f "$1" ]; then
					echo "$1 is not a valid file."
					cleanup
					exit 1
				fi
				# structure tests allow specifying any number of configs,
				# which can live anywhere on the host file system. to simplify
				# the docker volume mount, we copy all of these configs into
				# a /tmp directory and mount this single directory into the
				# test image. this directory is cleaned up after testing.
				filename=$(basename "$1")
				cp "$1" "$CONFIG_DIR"/"$filename"
				CMD_STRING=$CMD_STRING" --config /cfg/$filename"
			fi
			shift
			;;
		*)
			usage
			;;
	esac
done

if [ -z "$IMAGE_NAME" ]; then
	usage
fi

if [ $VERBOSE -eq 1 ]; then
	CMD_STRING=$CMD_STRING" -test.v"
fi

if [ $PULL -eq 1 ]; then
	docker pull "$ST_IMAGE"
fi

docker rm st_container > /dev/null 2>&1 || true # remove container if already there
docker run -d --entrypoint="/bin/sh" --name st_container "$ST_IMAGE" > /dev/null 2>&1

# shellcheck disable=SC2086
docker run --rm --entrypoint="$ENTRYPOINT" --volumes-from st_container -v "$CONFIG_DIR":/cfg "$IMAGE_NAME" $CMD_STRING

docker rm st_container > /dev/null 2>&1
cleanup
