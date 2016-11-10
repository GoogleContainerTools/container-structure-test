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

usage() {
	echo "Usage: $0 [-i <image>] [-c <config>] [-v] [-e <entrypoint>]"
	exit 1
}

export DOCKER_API_VERSION="1.21"

export VERBOSE=0
export CMD_STRING="/workspace/structure_test"
export ENTRYPOINT="/bin/sh"

while test $# -gt 0; do
	case "$1" in
		--image|-i)
			shift
			if test $# -gt 0; then
				export IMAGE_NAME=$1
			else
				usage
			fi
			shift
			;;
		--verbose|-v)
			export VERBOSE=1
			shift
			;;
		--entrypoint|-e)
			shift
			if test $# -gt 0; then
				export ENTRYPOINT=$1
			else
				usage
			fi
			shift
			;;
		--config|-c)
			shift
			if test $# -eq 0; then
				usage
			else
				CMD_STRING=$CMD_STRING" --config $1"
			fi
			shift
			;;
		*)
			usage
			;;
	esac
done

cp /test/* /workspace/

if [ $VERBOSE -eq 1 ]; then
	CMD_STRING=$CMD_STRING" -test.v"
fi

if [ -z "$IMAGE_NAME" ]; then
	usage
fi

docker run --privileged=true -v /workspace:/workspace --entrypoint="$ENTRYPOINT" "$IMAGE_NAME" -c "$CMD_STRING"
