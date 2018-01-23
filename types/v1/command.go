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

	"github.com/GoogleCloudPlatform/container-structure-test/types/unversioned"
)

type CommandTest struct {
	Name           string               `yaml:"name"`
	Setup          [][]string           `yaml:"setup"`
	Teardown       [][]string           `yaml:"teardown"`
	EnvVars        []unversioned.EnvVar `yaml:"envVars"`
	ExitCode       int                  `yaml:"exitCode"`
	Command        []string             `yaml:"command"`
	ExpectedOutput []string             `yaml:"expectedOutput"`
	ExcludedOutput []string             `yaml:"excludedOutput"`
	ExpectedError  []string             `yaml:"expectedError"`
	ExcludedError  []string             `yaml:"excludedError" ` // excluded error from running command
}

func validateCommandTest(t *testing.T, tt CommandTest) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Command == nil || len(tt.Command) == 0 {
		t.Fatalf("Please provide a valid command to run for test %s", tt.Name)
	}
	if tt.Setup != nil {
		for _, c := range tt.Setup {
			if len(c) == 0 {
				t.Fatalf("Error in setup command configuration encountered; please check formatting and remove all empty setup commands.")
			}
		}
	}
	if tt.Teardown != nil {
		for _, c := range tt.Teardown {
			if len(c) == 0 {
				t.Fatalf("Error in teardown command configuration encountered; please check formatting and remove all empty teardown commands.")
			}
		}
	}
	if tt.EnvVars != nil {
		for _, envVar := range tt.EnvVars {
			if envVar.Key == "" || envVar.Value == "" {
				t.Fatalf("Please provide non-empty keys and values for all specified env vars")
			}
		}
	}
}

func (ct CommandTest) LogName() string {
	return fmt.Sprintf("Command Test: %s", ct.Name)
}
