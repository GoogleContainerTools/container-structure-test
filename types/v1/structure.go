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
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/runtimes-common/structure_tests/drivers"
	"github.com/GoogleCloudPlatform/runtimes-common/structure_tests/types/unversioned"
	"github.com/GoogleCloudPlatform/runtimes-common/structure_tests/utils"
)

type StructureTest struct {
	DriverImpl         func(string) drivers.Driver
	Image              string
	GlobalEnvVars      []unversioned.EnvVar
	CommandTests       []CommandTest
	FileExistenceTests []FileExistenceTest
	FileContentTests   []FileContentTest
	LicenseTests       []LicenseTest
}

func (st *StructureTest) NewDriver() drivers.Driver {
	return st.DriverImpl(st.Image)
}

func (st *StructureTest) SetDriverImpl(f func(string) drivers.Driver, image string) {
	st.DriverImpl = f
	st.Image = image
}

func (st *StructureTest) RunAll(t *testing.T) int {
	testsRun := 0
	testsRun += st.RunCommandTests(t)
	testsRun += st.RunFileExistenceTests(t)
	testsRun += st.RunFileContentTests(t)
	testsRun += st.RunLicenseTests(t)
	return testsRun
}

func (st *StructureTest) RunCommandTests(t *testing.T) int {
	counter := 0
	for _, tt := range st.CommandTests {
		t.Run(tt.LogName(), func(t *testing.T) {
			validateCommandTest(t, tt)
			driver := st.NewDriver()
			vars := append(st.GlobalEnvVars, tt.EnvVars...)
			driver.Setup(t, vars, tt.Setup)

			stdout, stderr, exitcode := driver.ProcessCommand(t, tt.EnvVars, tt.Command)

			CheckOutput(t, tt, stdout, stderr, exitcode)
			counter++
		})
	}
	return counter
}

func (st *StructureTest) RunFileExistenceTests(t *testing.T) int {
	counter := 0
	for _, tt := range st.FileExistenceTests {
		t.Run(tt.LogName(), func(t *testing.T) {
			validateFileExistenceTest(t, tt)
			driver := st.NewDriver()
			var err error
			var info os.FileInfo
			info, err = driver.StatFile(t, tt.Path)
			if tt.ShouldExist && err != nil {
				t.Errorf("File %s should exist but does not!", tt.Path)
			} else if !tt.ShouldExist && err == nil {
				t.Errorf("File %s should not exist but does!", tt.Path)
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

func (st StructureTest) RunFileContentTests(t *testing.T) int {
	counter := 0
	for _, tt := range st.FileContentTests {
		t.Run(tt.LogName(), func(t *testing.T) {
			validateFileContentTest(t, tt)
			driver := st.NewDriver()
			actualContents, err := driver.ReadFile(t, tt.Path)
			if err != nil {
				t.Errorf("Failed to open %s. Error: %s", tt.Path, err)
			}

			contents := string(actualContents[:])

			var errMessage string
			for _, s := range tt.ExpectedContents {
				errMessage = "Expected string " + s + " not found in file contents!"
				utils.CompileAndRunRegex(s, contents, t, errMessage, true)
			}
			for _, s := range tt.ExcludedContents {
				errMessage = "Excluded string " + s + " found in file contents!"
				utils.CompileAndRunRegex(s, contents, t, errMessage, false)
			}
			counter++
		})
	}
	return counter
}

func (st *StructureTest) RunLicenseTests(t *testing.T) int {
	for num, tt := range st.LicenseTests {
		t.Run(tt.LogName(num), func(t *testing.T) {
			driver := st.NewDriver()
			checkLicenses(t, tt, driver)
		})
		return 1
	}
	return 0
}

func CheckOutput(t *testing.T, tt CommandTest, stdout string, stderr string, exitCode int) {
	for _, errStr := range tt.ExpectedError {
		errMsg := fmt.Sprintf("Expected string '%s' not found in error!", errStr)
		utils.CompileAndRunRegex(errStr, stderr, t, errMsg, true)
	}
	for _, errStr := range tt.ExcludedError {
		errMsg := fmt.Sprintf("Excluded string '%s' found in error!", errStr)
		utils.CompileAndRunRegex(errStr, stderr, t, errMsg, false)
	}
	for _, outStr := range tt.ExpectedOutput {
		errMsg := fmt.Sprintf("Expected string '%s' not found in output!", outStr)
		utils.CompileAndRunRegex(outStr, stdout, t, errMsg, true)
	}
	for _, outStr := range tt.ExcludedOutput {
		errMsg := fmt.Sprintf("Excluded string '%s' found in output!", outStr)
		utils.CompileAndRunRegex(outStr, stdout, t, errMsg, false)
	}
	if tt.ExitCode != exitCode {
		t.Errorf("Test %s exited with incorrect error code! Expected: %d, Actual: %d", tt.Name, tt.ExitCode, exitCode)
	}
}
