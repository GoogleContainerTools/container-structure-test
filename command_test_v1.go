// Copyright 2016 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"testing"
)

type CommandTestv1 struct {
	Name           string
	Setup          []Command
	Teardown       []Command
	EnvVars        []EnvVar
	ExitCode       int
	Command        []string
	ExpectedOutput []string
	ExcludedOutput []string
	ExpectedError  []string
	ExcludedError  []string // excluded error from running command
}

func validateCommandTestV1(t *testing.T, tt CommandTestv1) {
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
		for _, env_var := range tt.EnvVars {
			if env_var.Key == "" || env_var.Value == "" {
				t.Fatalf("Please provide non-empty keys and values for all specified env_vars")
			}
		}
	}
}

func (ct CommandTestv1) Name() string {
	return fmt.Sprintf("Command Test: %s", ct.Name)
}
