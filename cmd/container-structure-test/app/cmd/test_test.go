// Copyright 2018 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import "testing"

func TestSplitImagePath(t *testing.T) {
	tables := []struct {
		path string
		name string
		tag  string
		len  int
	}{
		{"image", "image", "", 1},
		{"image:tag", "image", "tag", 2},
		{"path/image", "path/image", "", 1},
		{"path/image:tag", "path/image", "tag", 2},
		{"my.registry:50000/path/image", "my.registry:50000/path/image", "", 1},
		{"my.registry:50000/path/image:tag", "my.registry:50000/path/image", "tag", 2},
		{"gcr.io/dga-demo/skaffold-example@sha256:44092b2ea3da5b9adc3c51c2ff6b399ae487094183a3746dbb8918d450d52ac5", "gcr.io/dga-demo/skaffold-example", "sha256:44092b2ea3da5b9adc3c51c2ff6b399ae487094183a3746dbb8918d450d52ac5", 2},
		{"gcr.io/dga-demo/skaffold-example:96be410b-dirty@sha256:44092b2ea3da5b9adc3c51c2ff6b399ae487094183a3746dbb8918d450d52ac5", "gcr.io/dga-demo/skaffold-example:96be410b-dirty", "sha256:44092b2ea3da5b9adc3c51c2ff6b399ae487094183a3746dbb8918d450d52ac5", 2},
	}

	for _, table := range tables {
		parts := splitImagePath(table.path)
		if parts[0] != table.name && parts[1] != table.tag && len(parts) != table.len {
			t.Errorf("Splitting image path (%v) was incorrect, got: %v:%v, expected: %v:%v", table.path, parts[0], parts[1], table.name, table.tag)
		}
	}
}
