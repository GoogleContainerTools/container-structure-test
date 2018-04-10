#!/bin/bash
# shellcheck disable=SC2181
# shellcheck disable=SC2046
files=$(buildifier -mode=check $(find . -not -path "./vendor/*" -name 'BUILD' -o -name '*.bzl' -type f))

if [[ $files ]]; then
  echo "$files"
  echo "Run 'buildifier -mode fix \$(find . -name BUILD -o -name '*.bzl' -type f)' to fix formatting"
  exit 1
fi
