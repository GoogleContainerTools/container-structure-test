#!/bin/sh

# Copyright 2024 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eu


LATEST=$(curl --silent -L https://api.github.com/repos/GoogleContainerTools/container-structure-test/releases/latest | jq -r .tag_name)

TAG="${TAG:-$LATEST}"


echo "Using tag: $LATEST"

checksums="$(curl --silent -L https://github.com/GoogleContainerTools/container-structure-test/releases/download/$TAG/checksums.txt)"


echo "Paste this into repositories.bzl"
echo ""
echo ""
echo ""

echo "_VERSION=\"$TAG\""
echo "_HASHES = {"
while IFS= read -r line; do
	read -r sha256 filename <<< "$line"
	integrity="sha256-$(echo $sha256 | xxd -r -p | base64)"
	filename=${filename#container-structure-test-}
    echo "    \"$filename\": \"$integrity\""
done <<< "$checksums"
echo "}"

echo ""
