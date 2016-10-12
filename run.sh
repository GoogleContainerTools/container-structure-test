#!/bin/sh

export DOCKER_API_VERSION="1.21"

IMAGE_NAME=$1
if [ -z "$1" ]; then
  echo "Please provide fully qualified path to image under test."
  exit 1
fi

cp /test/* /workspace/

docker run --privileged=true -v /workspace:/workspace "$IMAGE_NAME" /bin/sh -c "/workspace/structure_test"
