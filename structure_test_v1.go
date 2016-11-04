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
		var cmd *exec.Cmd
		if tt.Flags != nil && len(tt.Flags) > 0 {
			cmd = exec.Command(tt.Command, tt.Flags...)
		} else {
			cmd = exec.Command(tt.Command)
		}
		t.Logf("Executing: %s", cmd.Args)
		var outbuf, errbuf bytes.Buffer

		cmd.Stdout = &outbuf
		cmd.Stderr = &errbuf

		if err := cmd.Run(); err != nil {
			// The test might be designed to run a command that exits with an error.
			t.Logf("Error running command: %s. Continuing.", err)
		}

		stdout := outbuf.String()
		if stdout != "" {
			t.Logf("stdout: %s", stdout)
		}
		stderr := errbuf.String()
		if stderr != "" {
			t.Logf("stderr: %s", stderr)
		}

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
			t.Errorf("File %s should exist but does not!", tt.Path)
		} else if !tt.ShouldExist && err == nil {
			t.Errorf("File %s should not exist but does!", tt.Path)
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
