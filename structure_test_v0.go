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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

type StructureTestv0 struct {
	GlobalEnvVars      []EnvVar
	CommandTests       []CommandTestv0
	FileExistenceTests []FileExistenceTestv0
	FileContentTests   []FileContentTestv0
	LicenseTests       []LicenseTestv0
}

func (st StructureTestv0) RunAll(t *testing.T) int {
	originalVars := SetEnvVars(t, st.GlobalEnvVars)
	defer ResetEnvVars(t, originalVars)
	testsRun := 0
	testsRun += st.RunCommandTests(t)
	testsRun += st.RunFileExistenceTests(t)
	testsRun += st.RunFileContentTests(t)
	testsRun += st.RunLicenseTests(t)
	return testsRun
}

func (st StructureTestv0) RunCommandTests(t *testing.T) int {
	counter := 0
	for _, tt := range st.CommandTests {
		t.Run(tt.LogName(), func(t *testing.T) {
			validateCommandTestv0(t, tt)
			for _, setup := range tt.Setup {
				ProcessCommand(t, tt.EnvVars, setup, tt.ShellMode, false)
			}

			stdout, stderr, exitcode := ProcessCommand(t, tt.EnvVars, tt.Command, tt.ShellMode, true)
			CheckOutputv0(t, tt, stdout, stderr, exitcode)

			for _, teardown := range tt.Teardown {
				ProcessCommand(t, tt.EnvVars, teardown, tt.ShellMode, false)
			}
			counter++
		})
	}
	return counter
}

func (st StructureTestv0) RunFileExistenceTests(t *testing.T) int {
	counter := 0
	for _, tt := range st.FileExistenceTests {
		t.Run(tt.LogName(), func(t *testing.T) {
			validateFileExistenceTestv0(t, tt)
			var err error
			var info os.FileInfo
			if tt.IsDirectory {
				info, err = os.Stat(tt.Path)
			} else {
				info, err = os.Stat(tt.Path)
			}
			if tt.ShouldExist && err != nil {
				if tt.IsDirectory {
					t.Errorf("Directory %s should exist but does not!", tt.Path)
				} else {
					t.Errorf("File %s should exist but does not!", tt.Path)
				}
			} else if !tt.ShouldExist && err == nil {
				if tt.IsDirectory {
					t.Errorf("Directory %s should not exist but does!", tt.Path)
				} else {
					t.Errorf("File %s should not exist but does!", tt.Path)
				}
			}
			if tt.Permissions != "" {
				perms := info.Mode()
				if perms.String() != tt.Permissions {
					t.Errorf("%s has incorrect permissions. Expected: %s, Actual: %s", tt.Path, tt.Permissions, perms.String())
				}
			}
			counter++
		})
	}
	return counter
}

func (st StructureTestv0) RunFileContentTests(t *testing.T) int {
	counter := 0
	for _, tt := range st.FileContentTests {
		t.Run(tt.LogName(), func(t *testing.T) {
			validateFileContentTestv0(t, tt)
			actualContents, err := ioutil.ReadFile(tt.Path)
			if err != nil {
				t.Errorf("Failed to open %s. Error: %s", tt.Path, err)
			}

			contents := string(actualContents[:])

			var errMessage string
			for _, s := range tt.ExpectedContents {
				errMessage = "Expected string " + s + " not found in file contents!"
				compileAndRunRegex(s, contents, t, errMessage, true)
			}
			for _, s := range tt.ExcludedContents {
				errMessage = "Excluded string " + s + " found in file contents!"
				compileAndRunRegex(s, contents, t, errMessage, false)
			}
			counter++
		})
	}
	return counter
}

func (st StructureTestv0) RunLicenseTests(t *testing.T) int {
	for num, tt := range st.LicenseTests {
		t.Run(tt.LogName(num), func(t *testing.T) {
			checkLicenses(t, tt)
		})
		return 1
	}
	return 0
}

func CheckOutputv0(t *testing.T, tt CommandTestv0, stdout string, stderr string, exitCode int) {
	for _, errStr := range tt.ExpectedError {
		errMsg := fmt.Sprintf("Expected string '%s' not found in error!", errStr)
		compileAndRunRegex(errStr, stderr, t, errMsg, true)
	}
	for _, errStr := range tt.ExcludedError {
		errMsg := fmt.Sprintf("Excluded string '%s' found in error!", errStr)
		compileAndRunRegex(errStr, stderr, t, errMsg, false)
	}
	for _, outStr := range tt.ExpectedOutput {
		errMsg := fmt.Sprintf("Expected string '%s' not found in output!", outStr)
		compileAndRunRegex(outStr, stdout, t, errMsg, true)
	}
	for _, outStr := range tt.ExcludedOutput {
		errMsg := fmt.Sprintf("Excluded string '%s' found in output!", outStr)
		compileAndRunRegex(outStr, stdout, t, errMsg, false)
	}
	if tt.ExitCode != exitCode {
		t.Errorf("Test %s exited with incorrect error code! Expected: %d, Actual: %d", tt.Name, tt.ExitCode, exitCode)
	}
}
