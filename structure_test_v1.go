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
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"
)

type StructureTestv1 struct {
	CommandTests       []CommandTestv1
	FileExistenceTests []FileExistenceTestv1
	FileContentTests   []FileContentTestv1
}

func (st StructureTestv1) RunAll(t *testing.T) {
	st.RunCommandTests(t)
	st.RunFileExistenceTests(t)
	st.RunFileContentTests(t)
}

func (st StructureTestv1) RunCommandTests(t *testing.T) {
	for _, tt := range st.CommandTests {
		validateCommandTestV1(t, tt)
		for _, setup := range tt.Setup {
			ProcessCommand(t, setup, false)
		}

		stdout, stderr := ProcessCommand(t, tt.Command, true)
		CheckOutput(t, tt, stdout, stderr)

		for _, teardown := range tt.Teardown {
			ProcessCommand(t, teardown, false)
		}
	}
}

func (st StructureTestv1) RunFileExistenceTests(t *testing.T) {
	for _, tt := range st.FileExistenceTests {
		validateFileExistenceTestV1(t, tt)
		var err error
		if tt.IsDirectory {
			_, err = ioutil.ReadDir(tt.Path)
		} else {
			_, err = ioutil.ReadFile(tt.Path)
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
	}
}

func (st StructureTestv1) RunFileContentTests(t *testing.T) {
	for _, tt := range st.FileContentTests {
		validateFileContentTestV1(t, tt)
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
	}
}

func ProcessCommand(t *testing.T, fullCommand []string, checkOutput bool) (string, string) {
	var cmd *exec.Cmd
	command := fullCommand[0]
	flags := fullCommand[1:]
	if len(flags) > 0 {
		cmd = exec.Command(command, flags...)
	} else {
		cmd = exec.Command(command)
	}

	if checkOutput {
		t.Logf("Executing: %s", cmd.Args)
	} else {
		t.Logf("Executing setup/teardown: %s", cmd.Args)
	}

	var outbuf, errbuf bytes.Buffer

	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout := outbuf.String()
	if stdout != "" {
		t.Logf("stdout: %s", stdout)
	}
	stderr := errbuf.String()
	if stderr != "" {
		t.Logf("stderr: %s", stderr)
	}
	if err != nil {
		if checkOutput {
			// The test might be designed to run a command that exits with an error.
			t.Logf("Error running command: %s. Continuing.", err)
		} else {
			t.Fatalf("Error running setup/teardown command: %s.", err)
		}
	}
	return stdout, stderr
}

func CheckOutput(t *testing.T, tt CommandTestv1, stdout string, stderr string) {
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
}
