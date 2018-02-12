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
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/container-structure-test/drivers"
	"github.com/GoogleCloudPlatform/container-structure-test/types/unversioned"
	"github.com/GoogleCloudPlatform/container-structure-test/utils"
)

type StructureTest struct {
	DriverImpl         func(drivers.DriverConfig) (drivers.Driver, error)
	DriverArgs         drivers.DriverConfig
	GlobalEnvVars      []unversioned.EnvVar `yaml:"globalEnvVars"`
	CommandTests       []CommandTest        `yaml:"commandTests"`
	FileExistenceTests []FileExistenceTest  `yaml:"fileExistenceTests"`
	FileContentTests   []FileContentTest    `yaml:"fileContentTests"`
	MetadataTest       MetadataTest         `yaml:"metadataTest"`
	LicenseTests       []LicenseTest        `yaml:"licenseTests"`
}

func (st *StructureTest) NewDriver() (drivers.Driver, error) {
	return st.DriverImpl(st.DriverArgs)
}

func (st *StructureTest) SetDriverImpl(f func(drivers.DriverConfig) (drivers.Driver, error), args drivers.DriverConfig) {
	st.DriverImpl = f
	st.DriverArgs = args
}

func (st *StructureTest) RunAll(t *testing.T) int {
	testsRun := 0
	testsRun += st.RunCommandTests(t)
	testsRun += st.RunFileExistenceTests(t)
	testsRun += st.RunFileContentTests(t)
	testsRun += st.RunMetadataTests(t)
	testsRun += st.RunLicenseTests(t)
	return testsRun
}

func (st *StructureTest) RunCommandTests(t *testing.T) int {
	counter := 0
	for _, tt := range st.CommandTests {
		t.Run(tt.LogName(), func(t *testing.T) {
			validateCommandTest(t, tt)
			//each test needs to have its own instance of the driver: create that here
			driver, err := st.NewDriver()
			if err != nil {
				t.Fatalf(err.Error())
			}
			defer driver.Destroy(t)
			vars := append(st.GlobalEnvVars, tt.EnvVars...)
			driver.Setup(t, vars, tt.Setup)

			fullCommand := append([]string{tt.Command}, tt.Args...)
			stdout, stderr, exitcode := driver.ProcessCommand(t, tt.EnvVars, fullCommand)

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
			driver, err := st.NewDriver()
			if err != nil {
				t.Fatalf(err.Error())
			}
			defer driver.Destroy(t)
			var info os.FileInfo
			info, err = driver.StatFile(t, tt.Path)
			if tt.ShouldExist && err != nil {
				t.Fatalf("File %s should exist but does not, got error: %s!", tt.Path, err)
			} else if !tt.ShouldExist && err == nil {
				t.Fatalf("File %s should not exist but does!", tt.Path)
			}
			if tt.Permissions != "" {
				perms := info.Mode()
				if perms.String() != tt.Permissions {
					t.Fatalf("%s has incorrect permissions. Expected: %s, Actual: %s", tt.Path, tt.Permissions, perms.String())
				}
			}
			counter++
		})
	}
	return counter
}

func (st *StructureTest) RunFileContentTests(t *testing.T) int {
	counter := 0
	for _, tt := range st.FileContentTests {
		t.Run(tt.LogName(), func(t *testing.T) {
			validateFileContentTest(t, tt)
			driver, err := st.NewDriver()
			if err != nil {
				t.Fatalf(err.Error())
			}
			defer driver.Destroy(t)
			actualContents, err := driver.ReadFile(t, tt.Path)
			if err != nil {
				t.Fatalf("Failed to open %s. Error: %s", tt.Path, err)
			}

			contents := string(actualContents[:])

			var errMessage string
			for _, s := range tt.ExpectedContents {
				errMessage = fmt.Sprintf(
					"Expected string: %s not found in file content string: %s!",
					s, contents)
				utils.CompileAndRunRegex(s, contents, t, errMessage, true)
			}
			for _, s := range tt.ExcludedContents {
				errMessage = fmt.Sprintf(
					"Excluded string: %s found in file content string: %s!",
					s, contents)
				utils.CompileAndRunRegex(s, contents, t, errMessage, false)
			}
			counter++
		})
	}
	return counter
}

func (st *StructureTest) RunMetadataTests(t *testing.T) int {
	t.Run(st.MetadataTest.LogName(), func(t *testing.T) {
		driver, err := st.NewDriver()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer driver.Destroy(t)
		imageConfig, err := driver.GetConfig(t)
		if err != nil {
			t.Fatalf(err.Error())
		}
		for _, pair := range st.MetadataTest.Env {
			if imageConfig.Env[pair.Key] == "" {
				t.Errorf("variable %s not found in image env", pair.Key)
			} else if imageConfig.Env[pair.Key] != pair.Value {
				t.Errorf("env var %s value does not match expected value: %s", pair.Key, pair.Value)
			}
		}

		if st.MetadataTest.Cmd != nil {
			if len(*st.MetadataTest.Cmd) != len(imageConfig.Cmd) {
				t.Errorf("Image Cmd %v does not match expected Cmd: %v", imageConfig.Cmd, *st.MetadataTest.Cmd)
			} else {
				for i := range *st.MetadataTest.Cmd {
					if (*st.MetadataTest.Cmd)[i] != imageConfig.Cmd[i] {
						t.Errorf("Image config Cmd does not match expected value: %s", *st.MetadataTest.Cmd)
					}
				}
			}
		}

		if st.MetadataTest.Entrypoint != nil {
			if len(*st.MetadataTest.Entrypoint) != len(imageConfig.Entrypoint) {
				t.Errorf("Image entrypoint %v does not match expected entrypoint: %v", imageConfig.Entrypoint, *st.MetadataTest.Entrypoint)
			} else {
				for i := range *st.MetadataTest.Entrypoint {
					if (*st.MetadataTest.Entrypoint)[i] != imageConfig.Entrypoint[i] {
						t.Errorf("Image config entrypoint does not match expected value: %s", *st.MetadataTest.Entrypoint)
					}
				}
			}
		}

		if st.MetadataTest.Workdir != "" && st.MetadataTest.Workdir != imageConfig.Workdir {
			t.Errorf("Image workdir %s does not match config workdir: %s", imageConfig.Workdir, st.MetadataTest.Workdir)
		}

		for _, port := range st.MetadataTest.ExposedPorts {
			if !valueInList(port, imageConfig.ExposedPorts) {
				t.Errorf("Port %s not found in config", port)
			}
		}

		for _, volume := range st.MetadataTest.Volumes {
			if !valueInList(volume, imageConfig.Volumes) {
				t.Errorf("Volume %s not found in config", volume)
			}
		}
	})
	return 1
}

func valueInList(target string, list []string) bool {
	for _, value := range list {
		if target == value {
			return true
		}
	}
	return false
}

func (st *StructureTest) RunLicenseTests(t *testing.T) int {
	for num, tt := range st.LicenseTests {
		t.Run(tt.LogName(num), func(t *testing.T) {
			driver, err := st.NewDriver()
			if err != nil {
				t.Fatalf(err.Error())
			}
			defer driver.Destroy(t)
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
