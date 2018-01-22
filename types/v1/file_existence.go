// Copyright 2017 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"fmt"
	"testing"
)

type FileExistenceTest struct {
	Name        string `yaml:"name"`        // name of test
	Path        string `yaml:"path"`        // file to check existence of
	ShouldExist bool   `yaml:"shouldExist"` // whether or not the file should exist
	Permissions string `yaml:"permissions"` // expected Unix permission string of the file, e.g. drwxrwxrwx
}

func validateFileExistenceTest(t *testing.T, tt FileExistenceTest) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Path == "" {
		t.Fatalf("Please provide a valid file path for test %s", tt.Name)
	}
}

func (ft FileExistenceTest) LogName() string {
	return fmt.Sprintf("File Existence Test: %s", ft.Name)
}
