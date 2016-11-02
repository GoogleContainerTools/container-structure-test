#!/bin/sh

VERBOSE=0
CMD_STRING=""
ENTRYPOINT="./test/structure_test"
ST_IMAGE="gcr.io/gcp-runtimes/structure_test"
CONFIG_COUNTER=0

CONFIG_DIR=$(pwd)/.cfg
mkdir -p "$CONFIG_DIR"

command -v docker > /dev/null 2>&1 || { echo "Docker is required to run GCP structure tests, but is not installed on this host."; exit 1; }
command docker ps > /dev/null 2>&1 || { echo "Cannot connect to the Docker daemon!"; exit 1; }

cleanup() {
	rm -rf "$CONFIG_DIR"
}

usage() {
	echo "Usage: $0 [-i <image>] [-c <config>] [-v] [-e <entrypoint>]"
	cleanup
	exit 1
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
				# structure tests allow specifying any number of json configs,
				# which can live anywhere on the host file system. to simplify
				# the docker volume mount, we copy all of these configs into
				# a /tmp directory and mount this single directory into the
				# test image. this directory is cleaned up after testing.
				cp "$1" "$CONFIG_DIR"/cfg_$CONFIG_COUNTER.json
				CMD_STRING=$CMD_STRING" --config /cfg/cfg_$CONFIG_COUNTER.json"
				CONFIG_COUNTER=$(( CONFIG_COUNTER + 1 ))
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

docker run -d --entrypoint="/bin/sh" --name st_container "$ST_IMAGE" > /dev/null 2>&1

# shellcheck disable=SC2086
docker run --rm --entrypoint="$ENTRYPOINT" --volumes-from st_container -v "$CONFIG_DIR":/cfg "$IMAGE_NAME" $CMD_STRING

docker rm st_container > /dev/null 2>&1
cleanup
