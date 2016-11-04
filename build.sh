#!/bin/bash

set -e

export VERSION=$1

if [ -z "$1" ]; then
  echo "Please provide valid version to tag image."
  exit 1
fi

envsubst < cloudbuild.yaml.in > cloudbuild.yaml
cd ..
gcloud alpha container builds create . --config=structure_tests/cloudbuild.yaml
