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

package v2

import (
	"fmt"
	"testing"
)

type FileContentTest struct {
	Name             string   `yaml:"name"`             // name of test
	Path             string   `yaml:"path"`             // file to check existence of
	ExpectedContents []string `yaml:"expectedContents"` // list of expected contents of file
	ExcludedContents []string `yaml:"excludedContents"` // list of excluded contents of file
}

func validateFileContentTest(t *testing.T, tt FileContentTest) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Path == "" {
		t.Fatalf("Please provide a valid file path for test %s", tt.Name)
	}
}

func (ft FileContentTest) LogName() string {
	return fmt.Sprintf("File Content Test: %s", ft.Name)
}
