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

	"github.com/sirupsen/logrus"

	"github.com/GoogleCloudPlatform/container-structure-test/pkg/drivers"
	types "github.com/GoogleCloudPlatform/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleCloudPlatform/container-structure-test/pkg/utils"
)

type FileContentTest struct {
	Name             string   `yaml:"name"`             // name of test
	Path             string   `yaml:"path"`             // file to check existence of
	ExpectedContents []string `yaml:"expectedContents"` // list of expected contents of file
	ExcludedContents []string `yaml:"excludedContents"` // list of excluded contents of file
}

func (ft FileContentTest) Validate() error {
	if ft.Name == "" {
		return fmt.Errorf("Please provide a valid name for every test")
	}
	if ft.Path == "" {
		return fmt.Errorf("Please provide a valid file path for test %s", ft.Name)
	}
	return nil
}

func (ft FileContentTest) LogName() string {
	return fmt.Sprintf("File Content Test: %s", ft.Name)
}

func (ft FileContentTest) Run(driver drivers.Driver) *types.TestResult {
	result := &types.TestResult{
		Name:   ft.LogName(),
		Pass:   true,
		Errors: make([]string, 0),
	}
	logrus.Info(ft.LogName())
	actualContents, err := driver.ReadFile(ft.Path)
	if err != nil {
		result.Errorf("Failed to open %s. Error: %s", ft.Path, err)
		result.Fail()
		return result
	}

	contents := string(actualContents)

	for _, s := range ft.ExpectedContents {
		if !utils.CompileAndRunRegex(s, contents, true) {
			result.Errorf("Expected string '%s' not found in file content string '%s'", s, contents)
			result.Fail()
		}
	}
	for _, s := range ft.ExcludedContents {
		if !utils.CompileAndRunRegex(s, contents, false) {
			result.Errorf("Excluded string '%s' found in file content string '%s'", s, contents)
			result.Fail()
		}
	}
	return result
}
