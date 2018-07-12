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

# Bump these on release
# These are only used for local builds, all released builds are done with Bazel
VERSION_MAJOR ?= 1
VERSION_MINOR ?= 3
VERSION_BUILD ?= 0

VERSION ?= v$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)

GOOS ?= $(shell go env GOOS)
GOARCH = amd64
PROJECT := container-structure-test
RELEASE_BUCKET ?= gcp-container-tools/structure-test

LD_FLAGS := -X github.com/GoogleContainerTools/container-structure-test/pkg/version.version=$(VERSION)

SUPPORTED_PLATFORMS := linux-$(GOARCH) darwin-$(GOARCH)

BUILD_DIR ?= ./out
BUCKET ?= structure-test
UPLOAD_LOCATION := gs://${BUCKET}

$(BUILD_DIR)/$(PROJECT): $(BUILD_DIR)/$(PROJECT)-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/$(PROJECT)-$(GOOS)-$(GOARCH) $@

$(BUILD_DIR)/$(PROJECT)-%-$(GOARCH): $(GO_FILES) $(BUILD_DIR)
	GOOS=$* GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags="$(LD_FLAGS)" -o $@ .

%.sha256: %
	shasum -a 256 $< &> $@

%.exe: %
	cp $< $@

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

.PRECIOUS: $(foreach platform, $(SUPPORTED_PLATFORMS), $(BUILD_DIR)/$(PROJECT)-$(platform))
.PHONY: cross
cross: $(foreach platform, $(SUPPORTED_PLATFORMS), $(BUILD_DIR)/$(PROJECT)-$(platform).sha256)

.PHONY: release
release: cross
	gsutil cp $(BUILD_DIR)/$(PROJECT)-* gs://$(RELEASE_BUCKET)/$(VERSION)/

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

image:
	docker build -t gcr.io/gcp-runtimes/container-structure-test:latest .