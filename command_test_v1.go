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
	"testing"
)

type CommandTestv1 struct {
	Name           string
	Command        string
	Flags          []string
	ExpectedOutput []string
	ExcludedOutput []string
	ExpectedError  []string
	ExcludedError  []string // excluded error from running command
}

func validateCommandTestV1(t *testing.T, tt CommandTestv1) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Command == "" {
		t.Fatalf("Please provide a valid command to run for test %s", tt.Name)
	}
	t.Logf("COMMAND TEST: %s", tt.Name)
}
